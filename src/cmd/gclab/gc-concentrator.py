from dataclasses import dataclass
import dataclasses
import re
import sys
import math
import random
from xml.sax.saxutils import escape

@dataclass(frozen=True, order=True)
class Bytes:
    bytes: float

    @classmethod
    def fromBits(cls, bits: int):
        if bits % 8 != 0:
            raise ValueError(f"{bits} not a multiple of 8")
        return Bytes(bits // 8)

    def __format__(self, format_spec):
        width, prec = 1, 1
        if format_spec:
            m = re.match(r"^([0-9]+)(?:\.([0-9]+))?$", format_spec)
            if not m:
                raise RuntimeError(f"bad format spec {format_spec!r}")
            width = int(m[1])
            if m[2]:
                prec = int(m[2])

        b = self.bytes
        widtha = max(width - 4, 1)
        if b >= 1 << 40:
            return "%*.*f TiB"% (widtha, prec, b / (1 << 40))
        if b >= 1 << 30:
            return "%*.*f GiB"% (widtha, prec, b / (1 << 30))
        if b >= 1 << 20:
            return "%*.*f MiB"% (widtha, prec, b / (1 << 20))
        if b >= 1 << 10:
            return "%*.*f KiB"% (widtha, prec, b / (1 << 10))
        widtha = max(width - 6, 1)
        return "%*d bytes" % (widtha, b)

    def __str__(self):
        return format(self)
    
    def __add__(self, o):
        if isinstance(o, Bytes):
            return Bytes(self.bytes + o.bytes)
        return NotImplemented
    
    def __sub__(self, o):
        if isinstance(o, Bytes):
            return Bytes(self.bytes - o.bytes)
        return NotImplemented
    
    def __mul__(self, o):
        if not isinstance(o, (int, float)):
            raise NotImplemented
        return Bytes(self.bytes * o)

    def __rmul__(self, o):
        return self.__mul__(o)

    def __floordiv__(self, o):
        if isinstance(o, Bytes):
            return self.bytes // o.bytes
        if isinstance(o, int):
            return Bytes(self.bytes // o)
        return NotImplemented

    def __truediv__(self, o):
        if isinstance(o, Bytes):
            return self.bytes / o.bytes
        if isinstance(o, (int, float)):
            return self.bytes / o
        return NotImplemented

    def __index__(self):
        return self.bytes

KiB = Bytes(1<<10)
MiB = Bytes(1<<20)
GiB = Bytes(1<<30)

@dataclass
class Config:
    heapBytes: Bytes
    Ps: int
    wordBytes: Bytes = Bytes(8)

def fixedBufNetwork(config: Config, fanOut=16, fanIn=8, bufBytes=4 * KiB):
    print(f"Fixed buffer network of {bufBytes} buffers, {fanOut=}, {fanIn=}")
    
    markRegionBytes = config.wordBytes * (1<<16) # 16 bit word offsets

    layers = []
    for segBytes, pSpan in bidiFan(config, markRegionBytes, fanOut, fanIn):
        layers.append(Layer(config, segBytes, pSpan, bufBytes))
    return Network(layers)

def densityNetwork(config: Config, oneBitPerNBytes=32, markRegionBytes=32 * MiB, fanOut=16, fanIn=4, overview=False):
    # The central idea here is that almost all of the overhead is from the
    # bottom layer, and that is directly proportional to the density we want to
    # achieve, so we can build a network that dials in a specific overhead by
    # setting the desired density.
    #
    # The overhead also directly controls how much work can be clogged up in
    # buffers before making it to the bitmap. We want to minimize that while
    # maximizing the efficiency of flushing to the bitmap.
    #
    # The size of the mark region (and the heap size) determines the number of
    # buffers at the bottom level. We want to minimize this to reduce the
    # overhead of tracking buffers, but maximize it (minimizing the size of each
    # buffer) to reduce contention.

    if overview:
        print(f"Density based network: base density 1 bit every {oneBitPerNBytes} bytes over {markRegionBytes} regions")

    # How big do the bottom buffers need to be to achieve the given density?
    #
    # oneBitPerNBytes = 1 / (8 * bufLen / (markRegionBytes / wordBytes))
    baseAddrBytes = addrBytes(config, markRegionBytes)
    oneBitPerNBits = 8 * oneBitPerNBytes
    bufLen = markRegionBytes // (oneBitPerNBits * config.wordBytes)
    bufBytes = bufLen * baseAddrBytes

    layers = []
    for segBytes, pSpan in bidiFan(config, markRegionBytes, fanOut, fanIn):
        # Using smaller buffers at the higher levels complicates things with no
        # appreciable benefit.
        #
        # thisBufBytes = 4<<10
        # if segBytes == markRegionBytes:
        #     thisBufBytes = bufBytes
        thisBufBytes = bufBytes
        layers.append(Layer(config, segBytes, pSpan, thisBufBytes))
    return Network(layers)

def bidiFan(config: Config, markRegionBytes: Bytes, fanOut: int, fanIn: int):
    # Go from the bottom up to figure out the heap segmentation
    heapSpan = [markRegionBytes]
    while heapSpan[-1] < config.heapBytes:
        heapSpan.append(heapSpan[-1] * fanOut)

    # Go from the top down to figure out P fan in
    pSpan = [1]
    while pSpan[-1] < config.Ps:
        pSpan.append(pSpan[-1] * fanIn)

    # Align the two fans.
    while len(heapSpan) < len(pSpan):
        heapSpan.append(heapSpan[-1])
    while len(pSpan) < len(heapSpan):
        pSpan.append(pSpan[-1])
    heapSpan.reverse()

    # max(ceil(log_fo(heap)), ceil(log_fi(Ps))) layers
    return zip(heapSpan, pSpan)

@dataclass
class Layer:
    config: Config
    segBytes: Bytes
    pSpan: int
    bufBytes: Bytes

    def numBufs(self) -> int:
        """The total number of buffers in this layer."""
        pCount = (self.config.Ps + self.pSpan - 1) // self.pSpan
        heapCount = (self.config.heapBytes + self.segBytes - Bytes(1)) // self.segBytes
        return pCount * heapCount

    def bufLen(self) -> int:
        """The number of addresses that can be stored in a buffer in this layer."""
        return self.bufBytes // addrBytes(self.config, self.segBytes)

class Network(object):
    def __init__(self, layers: list[Layer]):
        self.config = layers[0].config
        self.layers = layers

    def __iter__(self):
        return iter(self.layers)
    
    def totalBufs(self):
        return sum(l.numBufs() for l in self)
    
    def totalBufBytes(self):
        return sum((l.numBufs() * l.bufBytes for l in self), start=Bytes(0))

def addrBytes(config: Config, segBytes: Bytes) -> Bytes:
    """The number of bytes to represent an address in a given heap segment."""
    words = segBytes // config.wordBytes
    if words <= 1<<8:
        return Bytes(1)
    if words <= 1<<16:
        return Bytes(2)
    if words <= 1<<32:
        return Bytes(4)
    return Bytes(8)

def printNetwork(network : Network):
    heapBytes = Bytes(1)
    for layer in network:
        heapBytes = layer.config.heapBytes
        nBuf = layer.numBufs()
        words = layer.segBytes//layer.config.wordBytes
        print(f"{nBuf:6} buffers @ {layer.bufBytes:8}, each covering {layer.segBytes:8} ({words} words)")
        dartBoardBytes = Bytes(words/8)
        print(f"       {dartBoardBytes} dart board per region")
        # If this buffer is full, what's the average density of bits set in the dart board?
        density = layer.bufLen() / words
        print(f"       1/{1/density:g} full buffer dart board density = 1 bit every {Bytes.fromBits(1//density)}")
    print(f"{network.totalBufs():6} total buffers")
    totalBufBytes = network.totalBufBytes()
    print(f"{totalBufBytes} of buffers ({100 * totalBufBytes / heapBytes:.2}% overhead)")

class svgGen:
    def __init__(self, config):
        self.__body = []
        self.__pMap = self.__linearMap((0, config.Ps), (0, 200))
        self.__hMap = self.__linearMap((Bytes(0), config.heapBytes), (0, 400))

    @staticmethod
    def __linearMap(f, t):
        return lambda x: ((x-f[0])/(f[1]-f[0])) * (t[1]-t[0]) + t[0]

    def path(self, attrs: str):
        p = svgPath(attrs, self)
        self.__body.append(p)
        return p
    
    def pt(self, l, p, h):
        # Map l, p, and h into pixels
        lPix = l * 100 + 100
        pPix = self.__pMap(p)
        hPix = self.__hMap(h)
        return vec(lPix, pPix, hPix)

    def project(self, pt):
        sqrt2 = 2 ** 0.5
        # Project l, p, and h into x, y SVG coordinates
        x, y = (pt.p + 1.2*pt.h) / sqrt2 + pt.x, 0.5 * (pt.h - pt.p) / sqrt2 + pt.l + pt.y
        return vecXY(x, y), x+60, y

    def text(self, pt, text, baselineVec, anchor="middle"):
        _, x, y = self.project(pt)
        transform = f'translate({x},{y})'
        if baselineVec != None:
            _, zx, zy = self.project(vec())
            _, bx, by = self.project(baselineVec)
            skew = math.atan2(by-zy, bx-zx) / math.pi * 180
            transform += f' skewY({skew})'
        self.__body.append(f'<text font-family="Arial" font-size="12" transform="{transform}" text-anchor="{anchor}">{escape(text)}</text>')

    def rectXY(self, pt, w, h, attrs):
        _, x, y = self.project(pt)
        self.__body.append(f'<rect x="{x}" y="{y}" width="{w}" height="{h}" {attrs} />')

    def write(self, f):
        print('''<svg version="1.1" width="650" height="650" xmlns="http://www.w3.org/2000/svg">''', file=f)
        for line in self.__body:
            print('  %s' % line, file=f)
        print('''</svg>''', file=f)

@dataclass(frozen=True)
class vec:
    l: float = 0.0
    p: float = 0.0
    h: float = 0.0

    x: float = 0.0
    y: float = 0.0

    def __add__(self, o):
        return vec(self.l + o.l, self.p + o.p, self.h + o.h, self.x + o.x, self.y + o.y)

    def __sub__(self, o):
        return vec(self.l - o.l, self.p - o.p, self.h - o.h, self.x - o.x, self.y - o.y)
    
    def __neg__(self):
        return -1 * self

    def __mul__(self, v):
        return vec(self.l * v, self.p * v, self.h * v, self.x * v, self.y * v)

    def __rmul__(self, v):
        return vec(self.l * v, self.p * v, self.h * v, self.x * v, self.y * v)

    def dot(self, o):
        return self.l * o.l + self.p * o.p + self.h * o.h + self.x * o.x + self.y + o.y

    def len(self):
        return math.sqrt(self.l ** 2 + self.p ** 2 + self.h ** 2 + self.x ** 2 + self.y ** 2)
    
    def norm(self):
        return self * (1/self.len())

    def xy(self):
        if self.l != 0 or self.p != 0 or self.h != 0:
            raise ValueError("Non-XY vector")
        return self.x, self.y

vL = vec(l=1)
vP = vec(p=1)
vH = vec(h=1)

def vecXY(x, y):
    return vec(x=x, y=y)

class svgPath:
    def __init__(self, attrs: str, svgGen: svgGen):
        self.__attrs = attrs
        self.__g = svgGen
        self.__d = []
        self.__transform = lambda pt: pt
        self.__last = None
        self.xys = []

    def transform(self, t):
        self.__transform = t

    def M(self, pt: vec):
        return self.op("M", pt)

    def L(self, pt: vec):
        return self.op("L", pt)

    def Z(self):
        self.__d.append("Z")

    def l(self, *, l=None, p=None, h=None, delta=vec()):
        pt = self.__last
        print(pt, file=sys.stderr)
        pt = vec((l if l != None else pt.l),
                 (p if p != None else pt.p),
                 (h if h != None else pt.h)) + delta
        return self.L(pt)
    
    def arrow2(self, p1: vec, p2: vec, dir: vec):
        """Draw an arrow that would cap a box ending at p1, p2.
        The tip of the arrow will be the midpoint between p1 and p2."""
        w = (p2-p1).len()
        aw = w * 2.5
        tip = (p1 + p2) * 0.5
        dir = dir.norm()
        right = (p2 - p1).norm()
        self.L(p1 - aw/2*dir)
        self.L(tip - aw/2*dir - aw/2*right) # Leftmost point
        self.L(tip)
        self.L(tip - aw/2*dir + aw/2*right) # Rightmost point
        self.L(p2 - aw/2*dir)
        return self

    def startConnector(self, dir):
        """Start a connector from the current point in direction dir."""
        # Unused
        p1, _, _ = self.__g.project(self.__transform(self.__last))
        p2, _, _ = self.__g.project(self.__transform(self.__last + dir))
        self.__cLine = (p1, p2)
        return self.__cLine

    def endConnector(self, dir, pt):
        # Unused
        if self.__cLine == None:
            raise RuntimeError("endConnector without startConnector")
        p1, p2 = self.__cLine
        p3, _, _ = self.__g.project(self.__transform(pt-dir))
        p4, _, _ = self.__g.project(self.__transform(pt))
        elbow = lineLineIntersection(p1, p2, p3, p4)
        self.L(elbow)
        if pt == self.xys[0]:
            self.Z()
        else:
            self.L(pt)
        self.__cLine = None
    
    def arrow(self, tip, dir, right, w):
        # Unused
        aw = w * 2.5
        self.endConnector(dir, tip - aw/2*dir - w/2*right)
        self.L(tip - aw/2*dir - aw/2*right)
        self.L(tip)
        self.L(tip - aw/2*dir + aw/2*right)
        self.L(tip - aw/2*dir + w/2*right)
        self.startConnector(-dir)

    def op(self, op, pt: vec):
        self.__last = pt
        vec2D, x, y = self.__g.project(self.__transform(pt))
        self.xys.append(vec2D)
        self.__d.append(f"{op} {x},{y}")
        return self

    def __str__(self):
        return '<path d="%s" %s />' % (" ".join(self.__d), self.__attrs)

def lineLineIntersection(a1, a2, b1, b2):
    x1, y1 = a1.xy()
    x2, y2 = a2.xy()
    x3, y3 = b1.xy()
    x4, y4 = b2.xy()
    x = (((x1 * y2 - y1 * x2) * (x3 - x4) - (x1 - x2) * (x3 * y4 - y3 * x4)) /
         ((x1 - x2) * (y3 - y4) - (y1 - y2) * (x3 - x4)))
    y = (((x1 * y2 - y1 * x2) * (y3 - y4) - (y1 - y2) * (x3 * y4 - y3 * x4)) /
         ((x1 - x2) * (y3 - y4) - (y1 - y2) * (x3 - x4)))
    return vecXY(x, y)

def linePointDistance(a1, a2, p):
    return (abs((a2.y - a1.y) * p.x - (a2.x - a1.x) * p.y + a2.x * a1.y - a2.y * a1.x) /
            math.sqrt((a2.y - a1.y)**2 + (a2.x - a1.x)**2))

def svgNetwork(f, network : Network):
    margin = 5
    arrowColor = 'fill="#3465a4" stroke="#204a87"'
    textHeight = vec(y=15)

    config = network.config
    g = svgGen(config)
    pt = g.pt

    pSpans = [[(p, min(p+layer.pSpan, config.Ps)) 
               for p in range(0, config.Ps, layer.pSpan)]
        for layer in network]
    hSpans = [[(Bytes(h), min(Bytes(h)+layer.segBytes, config.heapBytes))
               for h in range(Bytes(0), config.heapBytes, layer.segBytes)]
        for layer in network]

    def containedIn(r, rs):
        ra, rb = r
        return [(a, b) for (a, b) in rs
                if ra <= a and b <= rb]
    def mid(a, b, by=0.5):
        return a * (1-by) + b * by
    def midH(l, p, hSpan):
        h1, h2 = hSpan
        return pt(l, p, mid(h1, h2))
    
    def drawHighlight(tl, br, attrs):
        pad = 3.5
        tl -= pad*(vH+vL)
        br += pad*(vH+vL)
        p = g.path(attrs)
        p.M(tl).L(vec(l=tl.l,p=tl.p,h=br.h)).L(br).L(vec(l=br.l,p=br.p,h=tl.h)).Z()

    def drawFan(fromPts, toPts, right, attrs):
        w = 3
        up, down = (-w)*vL, w*vL
        p = g.path(attrs)
        # Shift everything by half the margin.
        p.transform(lambda pt: pt-(margin/2)*right)
        # Start at the bottom left of the cross-bar. Go clockwise.
        cL = toPts[0].l - 20
        cPt = lambda pt: vec(cL, pt.p, pt.h) # Map a point to the center level.
        cMin = cPt(min(fromPts + toPts, key=lambda p: p.dot(right)))
        cMax = cPt(max(fromPts + toPts, key=lambda p: p.dot(right)))
        # Draw the left edge of the cross-bar.
        p.M(cMin+down-w*right).L(cMin+up-w*right)
        # Draw the "from" blocks.
        for fromPt in fromPts:
            p.L(cPt(fromPt)+up-w*right).L(fromPt-w*right).L(fromPt+w*right).L(cPt(fromPt)+up+w*right)
        # Draw the right edge of the cross-bar.
        p.L(cMax+up+w*right).L(cMax+down+w*right)
        # Draw the "to" blocks.
        for toPt in toPts[::-1]:
            # Top-right
            p.L(cPt(toPt)+down+w*right)
            # TODO: Use p.arrow
            # Down to right arrow flare
            aWidth = 8
            p.L(toPt+w*right-aWidth*vL).L(toPt+aWidth*(right-vL))
            # Arrow point
            p.L(toPt)
            # Up to left arrow flare
            p.L(toPt-aWidth*(right+vL)).L(toPt-w*right-aWidth*vL)
            # Up to cross-bar
            p.L(cPt(toPt)+down-w*right)

    def textSpan(text, l, pSpan, hSpan, baselineVec, adjust=vec(), **textArgs):
        if not isinstance(pSpan, tuple):
            pSpan = (pSpan, pSpan)
        if not isinstance(hSpan, tuple):
            hSpan = (hSpan, hSpan)
        center = pt(l, mid(*pSpan), mid(*hSpan)) + adjust
        # Adjust for margins.
        if pSpan[0] != pSpan[1]:
            center -= (margin/2)*vP
        if hSpan[0] != hSpan[1]:
            center -= (margin/2)*vH
        g.text(center, text, baselineVec, **textArgs)

    def labelLayer(pt, text):
        pxy, _, _ = g.project(pt)
        g.text(vecXY(490, pxy.y), text, None, "start")

    # TODO: Generate this at random.
    scanSrc = 4
    highlight = [0, 1, 4, 18]

    lastLayer = len(network.layers)-1

    # Draw heap
    heapSep = 50 * vL
    heapH = 20 * vL
    heapD = 20 * vP
    p = g.path('fill="#e9b96e" stroke="#c17d11"')
    p1 = pt(lastLayer, 0, Bytes(0)) + heapSep
    p2 = pt(lastLayer, 0, hSpans[-1][-1][-1]) + heapSep - margin*vH
    p.M(p1).L(p2).L(p2+heapH).L(p1+heapH).Z()
    p.M(p1).L(p2).L(p2+heapD).L(p1+heapD).Z()
    p.M(p2).L(p2+heapD).L(p2+heapD+heapH).L(p2+heapH).Z()
    #labelLayer(p2+heapH, "Heap")

    # Draw dartboard
    dCols, dRows = 2, 4      # "Bits" in bitmap block
    dSep = heapSep - 25 * vL # Separation from lastLayer
    dRand = random.Random()
    dRand.seed(0)
    for hi, hSpan in enumerate(hSpans[-1]):
        d1, d2 = pt(lastLayer, 0, hSpan[0]) + dSep, pt(lastLayer, 0, hSpan[1]) - margin*vH + dSep
        dWidth = (d2.h - d1.h) / dCols
        # Draw flush highlight
        if hi == highlight[-1]:
            drawHighlight(d1, d2 + dWidth*dRows*vL, 'fill="#729fcf" stroke="#3465a4"')
        # Draw scan highlight
        if hi == scanSrc:
            drawHighlight(d1, d2 + (heapSep - dSep + heapH), 'fill="#fcaf3e" stroke="#f57900"')
        # Draw bitmap block
        for i in range(dCols):
            for j in range(dRows):
                l, r = mid(d1, d2, i/dCols)+j*dWidth*vL, mid(d1, d2, (i+1)/dCols)+j*dWidth*vL
                fill = "#fff"
                if dRand.randrange(2) == 0:
                    fill = "#000"
                g.path(f'fill="{fill}" stroke="#888"').M(l).L(r).L(r+dWidth*vL).L(l+dWidth*vL).Z()
        # Draw arrow down to the dartboard
        if hi == highlight[-1]:
            drawFan([midH(lastLayer, 0, hSpan)], [midH(lastLayer, 0, hSpan) + dSep], vH, arrowColor)
    dBot = dSep + dWidth * dRows * vL
    # Label dartboard
    dBotRight = pt(lastLayer, 0, hSpans[-1][-1][-1]) + dBot
    dBotRight -= vecXY(0, 10)
    labelLayer(dBotRight, "Dartboard")
    labelLayer(dBotRight + textHeight, "1 bit / heap word")

    # Draw layers from the bottom up.
    for l in range(len(network.layers)-1, -1, -1):
        # Draw layer l.
        for pi, (p1, p2) in enumerate(pSpans[l]):
            for hi, (h1, h2) in enumerate(hSpans[l]):
                attrs = 'fill="#d3d7cf" stroke="#555753"'
                if pi == 0 and hi == highlight[l]:
                    attrs = 'fill="#729fcf" stroke="#3465a4"'
                g.path(attrs).M(pt(l,p1,h1)).L(pt(l,p1,h2)-margin*vH).L(pt(l,p2,h2)-margin*(vP+vH)).L(pt(l,p2,h1)-margin*vP).Z()

        # Label layer.
        labelLayer(pt(l, p2, h2), f'{len(pSpans[l])} \u00D7 {len(hSpans[l])}')
        labelLayer(pt(l, p2, h2) + textHeight, f'{network.layers[l].bufBytes:0.0} buffers')

        # Label heap span size.
        if h2-h1 > 256*MiB:
            text1, text2 = "", f"{h2-h1:0.0} span"
        else:
            text1, text2 = f"{h2-h1:0.0}", "span"
        p = g.path('fill="none" stroke="#000" stroke-width="2"')
        p1 = pt(l, pSpans[l][-1][-1], hSpans[l][-1][0]) - margin*vP
        p2 = pt(l, pSpans[l][-1][-1], hSpans[l][-1][-1]) - margin*(vP+vH)
        p.M(p1 - 3*vL)
        p.L(p1 - 8*vL)
        p.L(p2 - 8*vL)
        p.L(p2 - 3*vL)
        g.text(p2 - 8*vL - 5*vL, text2, vH, "end")
        if text1:
            g.text(p2 - 8*vL - 5*vL - textHeight, text1, vH, "end")

        #textSpan(text1, l, pSpans[l][-1][-1], hSpans[l][-1], vH, -textHeight, anchor="right")
        #textSpan(text2, l, pSpans[l][-1][-1], hSpans[l][-1], vH, vec(), anchor="right")

        if l == 0:
            continue

        # Draw the H fan from layer l up to layer l-1.
        for i, pSpan in enumerate(pSpans[l-1][::-1]):
            p = mid(*pSpan)
            for j, fromSpan in enumerate(hSpans[l-1]):
                toSpans = containedIn(fromSpan, hSpans[l])
                opacity = .2
                if i == len(pSpans[l-1]) - 1 and j == highlight[l-1]:
                    opacity = 1
                attrs = f'{arrowColor} opacity="{opacity}"'
                drawFan([midH(l-1, p, fromSpan)], [midH(l, p, toSpan) for toSpan in toSpans], vH, attrs)

    # Label axes
    # Label CPUs axis
    labelPos = -5*vL
    textSpan("CPU", 0, (pSpans[0][0][0], pSpans[0][-1][-1]), Bytes(0), vP, labelPos - 1*textHeight)
    for i, pSpan in enumerate(pSpans[0]):
        textSpan(f"{i}", 0, pSpan, Bytes(0), vP, labelPos)
    # Label heap axis
    ll = len(network.layers)-1
    fullHSpan = (hSpans[-1][0][0], hSpans[-1][-1][-1])
    # XXX Figure out baseline adjustment for real
    labelPos = heapSep + heapH + textHeight
    textSpan("0", ll, 0, fullHSpan[0], vH, labelPos)
    textSpan(f"{config.heapBytes:0.0}", ll, 0, fullHSpan[1], vH, labelPos - margin*vH)
    textSpan("\u22EF", ll, 0, fullHSpan, vH, labelPos - (margin/2)*vH)
    #textSpan("heap", ll, 0, fullHSpan, vH, labelPos + 0*textHeight - (margin/2)*vH)
    textSpan("Heap", ll, 0, fullHSpan, vH, heapSep + heapH - 5*vL - (margin/2)*vH)

    # Scan feedback loop. This is super tricky because we bounce between LVH
    # coordinates and XY coordinates.
    #
    # TODO: This still has a weird "twisting" effect. What does it look like if
    # we just put it all in the vH x vL plane?
    def drawFeedback(hSpan):
        arrowColor = 'fill="#f57900" stroke="#ce5c00"'
        textColor = 'fill="#fcaf3e" stroke="#f57900"'

        # Compute points where we enter and exit the concentrator network.        
        pTL = pt(0, 0, hSpan[0])
        pTR = pt(0, 0, hSpan[1]) - margin*vH
        pBL = pt(ll, 0, hSpan[0]) + heapSep + heapH
        pBR = pt(ll, 0, hSpan[1]) - margin*vH + heapSep + heapH

        # Draw bottom part
        p = g.path(arrowColor)
        # Start at the heap region
        offP = -120*vP
        p.M(pBL)
        p.L(pBR)
        p.L(pBR + offP)
        p.L(pBL + offP)

        # Compute position of arrow to label
        offL = -250*vL
        topLeft, _, _ = g.project(pBL + offP + offL)
        topRight, _, _ = g.project(pBR + offP + offL)
        y = (topLeft.y + topRight.y) / 2
        topLeft = dataclasses.replace(topLeft, y=y)
        topRight = dataclasses.replace(topRight, y=y)
        tip = 0.5 * (topLeft + topRight)

        # Draw label
        labelW = 35
        labelH = 20
        tip, _, _ = g.project(tip)
        g.rectXY(tip + vecXY(-labelW/2, -labelH), labelW, labelH, f'{textColor} rx="4"')
        g.text(tip + vec(y=-5), "scan", None)

        # Draw arrow up to label
        p = g.path(arrowColor)
        p.M(pBL + offP)
        p.arrow2(topLeft, topRight, vec(y=-1))
        p.L(pBR + offP)

        # Draw top part, starting at elbow
        p = g.path(arrowColor)
        p.M(pTR + offP)
        p.L(pTL + offP)
        p.arrow2(pTL, pTR, vP)
        p.Z()

        # Draw path up from label
        p = g.path(arrowColor)
        p.M(topRight - vecXY(0, labelH))
        p.L(topLeft - vecXY(0, labelH))
        p.L(pTL + offP)
        p.L(pTR + offP).Z()

    drawFeedback(hSpans[-1][scanSrc])

    g.write(f)

def expRange(start, n, exp=2):
    i = start
    for _ in range(n):
        yield i
        i *= exp

def mainRange():
    for Ps in [1<<n for n in range(10)]:
        heap = Ps * GiB  # 1 GiB per P
        #heap = 64 * MiB
        config = Config(heap, Ps)
        print(f"## {Ps} Ps, {heap} heap")
        printNetwork(fixedBufNetwork(config))
        print()
        printNetwork(densityNetwork(config, overview=True))
        print()

def mainOverheadGrid():
    #global Ps, heap
    heaps = list(expRange(64 * MiB, 10, 4))
    Pss = list(expRange(1, 10))

    print("heap \\ Ps".rjust(9), sep="", end="")
    for Ps in Pss:
        print(f"  {Ps:6}", sep="", end="")
    print()
    for heap in heaps:
        print(f"{heap:9}", sep="", end="")
        for Ps in Pss:
            config = Config(heap, Ps)
            network = densityNetwork(config)
            totalBufBytes = network.totalBufBytes()
            print(f" {100 * totalBufBytes / heap:6.2f}%", sep="", end="")
            #print(f" {totalBufBytes:7.0}", sep="", end="")
        print()

#mainRange()
#mainOverheadGrid()
#printNetwork(densityNetwork(Config(64 * GiB, 64), overview=True))

# inkscape -w 2048 -h 2048 /tmp/x.svg -b '#fff' -o x.png
svgNetwork(sys.stdout, densityNetwork(Config(1 * GiB, 8), markRegionBytes=32 * MiB, fanOut=4, fanIn=2))
#svgNetwork(sys.stdout, densityNetwork(Config(16 * GiB, 16), fanOut=16))
