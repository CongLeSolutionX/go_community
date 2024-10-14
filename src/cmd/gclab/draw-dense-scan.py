import os
import sys
import math
import random
from dataclasses import dataclass
from xml.sax.saxutils import escape

import seaborn as sns

@dataclass(frozen=True)
class vec:
    x: float = 0.0
    y: float = 0.0

    def __add__(self, o):
        return vec(self.x + o.x, self.y + o.y)

    def __sub__(self, o):
        return vec(self.x - o.x, self.y - o.y)
    
    def __neg__(self):
        return -1 * self

    def __mul__(self, v):
        return vec(self.x * v, self.y * v)

    def __rmul__(self, v):
        return vec(self.x * v, self.y * v)

    def dot(self, o):
        return self.x * o.x + self.y + o.y

    def len(self):
        return math.sqrt(self.x ** 2 + self.y ** 2)
    
    def norm(self):
        return self * (1/self.len())

class svgGen:
    def __init__(self, w, h):
        self.__defs = []
        self.__body = []
        self.__xMap = self.__linearMap((0, w), (10, 510))
        self.__yMap = self.__linearMap((0, h), (10, 510))

    @staticmethod
    def __linearMap(f, t):
        return lambda x: ((x-f[0])/(f[1]-f[0])) * (t[1]-t[0]) + t[0]

    def path(self, attrs: str):
        p = svgPath(attrs, self)
        self.__body.append(p)
        return p

    def project(self, vec: vec):
        return self.__xMap(vec.x), self.__yMap(vec.y)

    def text(self, pt, text, baselineVec=None, anchor="middle", attrs=""):
        x, y = self.project(pt)
        transform = f'translate({x},{y})'
        if baselineVec != None:
            zx, zy = self.project(vec())
            bx, by = self.project(baselineVec)
            skew = math.atan2(by-zy, bx-zx) / math.pi * 180
            transform += f' skewY({skew})'
        self.__body.append(f'<text font-family="Roboto Light" font-size="12" transform="{transform}" text-anchor="{anchor}" {attrs}>{escape(text)}</text>')

    def circle(self, pt, r, attrs):
        x, y = self.project(pt)
        dx, dy = self.project(pt + vec(r, r))
        self.__body.append(f'<circle cx="{x}" cy="{y}" r="{max(dx-x, dy-y)}" {attrs} />')

    def rect(self, rect, attrs, rx=0):
        x, y = self.project(rect.nw)
        x2, y2 = self.project(rect.se)
        self.__body.append(f'<rect x="{x}" y="{y}" width="{x2-x}" height="{y2-y}" rx="{rx}" {attrs} />')

    def linearGradient(self, name, vec, *stops):
        self.__defs.append(f'<linearGradient id="{escape(name)}" x1="0" y1="0" x2="{vec.x}" y2="{vec.y}">')
        for pos, color in stops:
            self.__defs.append(f'  <stop offset="{100*pos}%" stop-color="{color}" />')
        self.__defs.append(f'</linearGradient>')

    def write(self, f):
        print('''<svg version="1.1" width="520" height="410" xmlns="http://www.w3.org/2000/svg">''', file=f)
        if self.__defs:
            print('  <defs>', file=f)
            for line in self.__defs:
                print('    %s' % line, file=f)
            print('  </defs>', file=f)
        for line in self.__body:
            print('  %s' % line, file=f)
        print('''</svg>''', file=f)

class svgPath:
    def __init__(self, attrs: str, svgGen: svgGen):
        self.__attrs = attrs
        self.__g = svgGen
        self.__d = []
        self.__transform = lambda pt: pt
        self.__last = None

    def transform(self, t):
        self.__transform = t

    def M(self, pt: vec):
        return self.op("M", pt)

    def L(self, pt: vec):
        return self.op("L", pt)

    def Q(self, pt: vec, pt2: vec):
        return self.op2("Q", pt, pt2)

    def Z(self):
        self.__d.append("Z")

    def l(self, *, x=None, y=None, delta=vec()):
        pt = self.__last
        pt = vec((x if x != None else pt.x),
                 (y if y != None else pt.y)) + delta
        return self.L(pt)
    
    def rect(self, rect):
        return self.M(rect.nw).L(rect.ne).L(rect.se).L(rect.sw).Z()
    
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

    def op(self, op, pt: vec):
        self.__last = pt
        x, y = self.__g.project(self.__transform(pt))
        self.__d.append(f"{op} {x},{y}")
        return self

    def op2(self, op, pt: vec, pt2: vec):
        self.__last = pt2
        x, y = self.__g.project(self.__transform(pt))
        x2, y2 = self.__g.project(self.__transform(pt2))
        self.__d.append(f"{op} {x} {y},{x2} {y2}")
        return self

    def __str__(self):
        return '<path d="%s" %s />' % (" ".join(self.__d), self.__attrs)

@dataclass(frozen=True)
class rect:
    nw: vec
    se: vec

    @property
    def ne(self): return vec(self.se.x, self.nw.y)
    @property
    def sw(self): return vec(self.nw.x, self.se.y)
    @property
    def n(self): return vec(mid(self.nw.x, self.se.x), self.nw.y)
    @property
    def e(self): return vec(self.se.x, mid(self.nw.y, self.se.y))
    @property
    def s(self): return vec(mid(self.nw.x, self.se.x), self.se.y)
    @property
    def w(self): return vec(self.nw.x, mid(self.nw.y, self.se.y))

    def combine(self, o):
        return rect(vec(min(self.nw.x, o.nw.x), min(self.nw.y, o.nw.y)),
                    vec(max(self.se.x, o.se.x), max(self.se.y, o.se.y)))

def mid(a, b):
    return (a + b) / 2

def bbox(objs):
    if isinstance(objs, rect):
        return objs
    n = min(o.nw.y for o in objs)
    e = max(o.se.x for o in objs)
    s = max(o.se.y for o in objs)
    w = min(o.nw.x for o in objs)
    return rect(vec(w, n), vec(e, s))

#colors = ['#66c2a5','#fc8d62','#8da0cb','#e78ac3','#a6d854','#ffd92f','#e5c494','#b3b3b3']
#colors = sns.color_palette("deep")
#setColors = sns.color_palette("pastel")
colors = sns.color_palette("husl", 8+3)
colors2 = ["#729fcf", "#e9b96e"]

def rgb(r, g, b):
    return f"rgb({int(r*255):d},{int(g*255):d},{int(b*255):d})"

def main():
    g = svgGen(500, 500)
    pos = vec(0, 0)
    vZ = vec(5, -3)
    byteGap = 4

    # Dartboard random generator
    dRand = random.Random()
    dRand.seed(10)
    # Pointer mask random generator
    pRand = random.Random()
    pRand.seed(3)
    # TODO: Each constructed input should have an independent generator.

    spanWords = 1024
    objWords = 5
    leading = vec(y=40)

    bitCols = 64
    bitSize = 400 / bitCols

    # TODO: This would all be way less finicky if I could actually build up the
    # value/operation graph and then just tweak the positioning of nodes in that
    # and blat it out to SVG. Those operations could also compute the actual
    # results in addition to creating visual nodes. In particular, right now I
    # have to draw all of the inputs *and output* of an operation before I can
    # draw the operation so I can get all placement right, which is super
    # confusing.

    def bitmapHeight(bits, cols=bitCols):
        rows = (len(bits) + cols - 1) // cols
        return bitSize + (rows - 1) * -vZ.y
    
    def bitmapWidth(bits, cols=bitCols):
        # TODO: Maybe I should just compute the rectangles.
        n = min(cols, len(bits))
        return n * bitSize + (n - 1) // 8 * byteGap

    def drawBitmap2(bits, cols=bitCols, offX=0, dim=bitCols, bitWidth=bitSize):
        # Bottom left of the bitmap is at pos. (Layout is done like text.)
        off = vec(offX, -bitSize)
        rects = [None] * len(bits)
        for i, bit in reversed(list(enumerate(bits))):
            nw = pos + off + vec((i%cols) * bitWidth + (i%cols)//8 * byteGap, 0) + (i // cols) * vZ
            r = rect(nw, nw + vec(bitWidth, bitSize))
            if not bit:
                fill = "#fff"
            elif i >= dim:
                fill = "#999"
            else:
                fill = "#000"
            if i >= dim:
                stroke = "#ddd"
            else:
                stroke = "#888"
            g.path(f'fill="{fill}" stroke="{stroke}"').rect(r)
            rects[i] = r
        return rects

    def drawBusIn(srcs, colorNum=0):
        src = bbox(srcs)
        g.path(f'fill="none" stroke="{colors2[colorNum]}" stroke-width="1.5"').M(src.sw + vec(0, 3)).L(src.se + vec(0, 3))
        return src.s + vec(0, 3)

    def drawBusOut(dsts, colorNum=0):
        dst = bbox(dsts)
        g.path(f'fill="none" stroke="{colors2[colorNum]}" stroke-width="1.5"').M(dst.nw - vec(0, 3)).L(dst.ne - vec(0, 3))
        return dst.n - vec(0, 3)
    
    def drawBusElbow(a, b, c, colorNum=0, bendDistance=None, bendSize=5):
        path = g.path(f'fill="none" stroke="{colors2[colorNum]}" stroke-width="1.5"')
        path.M(a)
        if bendDistance is None:
            path.L(vec(a.x, b.y))
        else:
            inDir = vec(0, b.y).norm()
            p = a + inDir * bendDistance
            outDir = (b - p).norm()
            path.L(a + inDir * (bendDistance - bendSize))
            path.Q(p, p + outDir * bendSize)
        if c is None:
            path.L(b)
        else:
            path.L(vec(c.x, b.y)).L(c)

    def drawOp(src1, src2, op, dst, pos='c', offY=0, elbow=True):
        src1, src2, dst = bbox(src1), bbox(src2), bbox(dst)
        cy = mid(src2.s.y, dst.n.y) + offY
        p1, p2 = src1.s, src2.s # Source points 1 and 2
        off = vec(0, 0)
        if pos == 'c':
            off = vec(20, 0)
            p1, p2 = p1 - off, p2 + off
            cx = mid(p1.x, p2.x)
        elif pos == '1':
            cx = p1.x
        elif pos == '2':
            cx = p2.x
        c = vec(cx, cy) # Center
        if elbow:
            drawBusElbow(drawBusIn(src1), c, drawBusOut(dst))
            drawBusElbow(drawBusIn(src2, colorNum=1) - off, c, None, colorNum=1)
        else:
            drawBusElbow(drawBusIn(src1), c, drawBusOut(dst), bendDistance=5)
            drawBusElbow(drawBusIn(src2, colorNum=1) - off, c, None, colorNum=1, bendDistance=5)
        # Draw op at c
        g.circle(c, 10, 'fill="#fff" stroke="#000"')
        g.text(c, op, vec(1, 0), attrs='dominant-baseline="central"')
        
    def drawOp1(src, op, dst, opw):
        cy = bbox(src).s.y + leading.y * 0.5
        p1 = drawBusIn(src)
        p2 = drawBusOut(dst)
        c = vec(p2.x, cy)
        drawBusElbow(p1, c, p2)
        r = rect(c - vec(opw/2, 8), c + vec(opw/2, 8))
        g.rect(r, 'fill="#fff" stroke="#000"', rx="8")
        g.text(c, op, attrs='dominant-baseline="central"')

    g.linearGradient("zoom", vec(0, 1), (0, "#444"), (1, "#bbb"))
    def drawZoom(src, dst):
        g.path('fill="url(#zoom)" stroke="none"').M(src.sw).L(src.se).L(dst.ne).L(dst.nw).Z()

    def drawClusters(colors, opacity=[]):
        off = vec(0, -bitSize)
        rects = []
        for i, color in enumerate(colors):
            if color is None:
                continue
            nw = pos + off + vec(i * (8 * bitSize + byteGap), 0)
            r = rect(nw, nw + vec(8 * bitSize, bitSize))
            op = f' opacity="{opacity[i]}"' if i < len(opacity) else ""
            g.path(f'fill="{color}" stroke="#888"{op}').rect(r)
            rects.append(r)
        return rects

    # Generate dartboard
    endWord = spanWords // objWords * objWords
    # Higher chance in the first 64 bits
    dartboard = [w < endWord and dRand.randrange(8 if w < 64 else 16) == 0 for w in range(spanWords)]

    # Draw dartboard
    dRows = len(dartboard) // bitCols
    pos += vec(y=bitmapHeight(dartboard))
    rects = drawBitmap2(dartboard)
    fullWidth = bitmapWidth(dartboard)
    
    pos += leading

    # Focus on the first 64 words
    #
    # TODO: Draw full depth but have it fade away quickly to make it clear
    # there's more than what I'm showing on the slide?
    spanWords = 64
    dartboard = dartboard[:spanWords]
    endWord = spanWords
    rects = rects[:endWord]

    # Compute object darts
    objDarts = [False] * ((spanWords + objWords - 1) // objWords)
    for i, bit in enumerate(dartboard):
        if bit:
            objDarts[i//objWords] = True

    # Draw object darts
    centerX = fullWidth / 2
    objDartRects = drawBitmap2(objDarts, offX=centerX - bitmapWidth(objDarts)/2)

    # Connect dartboard to object darts
    m = vec(y=5)
    for i, r2 in enumerate(objDartRects):
        r1 = rects[i*objWords].combine(rects[min(len(rects)-1, (i+1)*objWords - 1)])
        g.path(f'fill="{colors2[i%len(colors2)]}" stroke="none"').M(r1.sw).L(r1.se).L(r1.se+m).L(r2.ne-m).L(r2.ne).L(r2.nw).L(r2.nw-m).L(r1.sw+m).Z()

    # Draw mark bits
    markBits = [False] * len(objDarts)
    for i, bit in enumerate(objDarts):
        if bit:
            p = 0.5
        else:
            p = 0.1
        markBits[i] = dRand.random() < p
    markBitsRects = drawBitmap2(markBits, offX=fullWidth - bitmapWidth(markBits))

    pos += leading * 1.5

    # Draw object scan bitmap
    objScan = [obj and not mark for obj, mark in zip(objDarts, markBits)]
    objScanRects = drawBitmap2(objScan, offX=centerX - bitmapWidth(objScan)/2)
    drawOp(objDartRects, markBitsRects, "&^", objScanRects, pos='1', elbow=False)

    # Draw updated mark bitmap
    newMarkBits = [obj or mark for obj, mark in zip(objDarts, markBits)]
    newMarkBitsRects = drawBitmap2(newMarkBits, offX=fullWidth - bitmapWidth(newMarkBits))
    drawOp(objDartRects, markBitsRects, "|", newMarkBitsRects, pos='2', elbow=False)

    pos += leading

    # Draw scan bitmap
    wordScan = [w < endWord and objScan[w//objWords] for w in range(spanWords)]
    wordScanRects = drawBitmap2(wordScan)
    pos += leading

    # Connect object darts to scan bitmap
    for i, r1 in enumerate(objScanRects):
        r2 = wordScanRects[i*objWords].combine(wordScanRects[min(len(rects)-1, (i+1)*objWords - 1)])
        g.path(f'fill="{colors2[i%len(colors2)]}" stroke="none"').M(r1.sw).L(r1.se).L(r1.se+m).L(r2.ne-m).L(r2.ne).L(r2.nw).L(r2.nw-m).L(r1.sw+m).Z()

    # Generate and draw pointer mask
    ptrMask = [pRand.randrange(3) == 0 for _ in range(spanWords)]
    ptrMaskRects = drawBitmap2(ptrMask)
    pos += leading * 1.5

    # AND pointer mask with scan bitmap
    wordScan = [s and p for s, p in zip(wordScan, ptrMask)]
    wordScan2Rects = drawBitmap2(wordScan)
    drawOp(wordScanRects, ptrMaskRects, '&', wordScan2Rects)
    pos += leading

    # Data from cluster 0
    dataColors = colors.copy()
    dRand.shuffle(dataColors)
    clusters = [rgb(*c) for c in dataColors[:8]]
    clustersRects = drawClusters(clusters)
    pos += leading * 1.5

    # Pack the masked words
    packedClusters = []
    # Lead this with some other words (faded) to show this as part of the buffer.
    packedOpacity = []
    for color in dataColors[8:]:
        packedClusters.append(rgb(*color))
        packedOpacity.append(0.5)
    packMap = {}
    for i, c in enumerate(clusters):
        if wordScan[i]:
            packMap[i] = len(packedClusters)
            packedClusters.append(c)
    packedClustersRects = drawClusters(packedClusters, packedOpacity)

    # Draw compress op
    cy = mid(bbox(clustersRects).s.y, bbox(packedClustersRects).n.y)
    for i, color in enumerate(clusters):
        inPos = clustersRects[i].s
        path = g.path(f'fill="none" stroke="{color}" stroke-width="3"').M(inPos).L(vec(inPos.x, cy+8))
        if i in packMap:
            inPos = packedClustersRects[packMap[i]].n
            path.L(inPos).L(inPos+vec(0,3))
    g.rect(rect(vec(0, cy-8), vec(fullWidth, cy+8)), 'fill="#fff" stroke="#000"', rx="8")
    maskPos = vec(0, cy-8) - vec(0, 5)
    g.text(vec(fullWidth/2, cy), "compress", attrs='dominant-baseline="central"')

    # Zoom in on the first 8 bits and show it as a mask for the compress
    #
    # TODO: I'm not very happy with this way of showing the mask input.
    pos = maskPos
    wordScanZoom = drawBitmap2(wordScan[:8], bitWidth=fullWidth/8)
    #drawZoom(bbox(wordScan2Rects[:8]), bbox(wordScanZoom))
    drawOp1(wordScan2Rects[:8], 'load mask', wordScanZoom, 70)

    g.write(sys.stdout)

main()
