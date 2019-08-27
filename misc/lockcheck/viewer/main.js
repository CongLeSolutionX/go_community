"use strict";

// TODO
//
// * Let user select between full graph and cycle graph. Maybe just an
//   option to hide excluded nodes and relayout.

function onload() {
    // Activate expand button.
    const graphView = document.getElementById("graph-view");
    const expand = document.getElementById("expand");
    expand.addEventListener("click", function(ev) {
        graphView.className = "full";
    });

    // Load the lock graph.
    loadGraph();
}

// fetchChecked wraps fetch, but shows a loading overlay and displays
// fetch errors to the user.
async function fetchChecked(url) {
    const elt = document.getElementById("loading");
    elt.innerText = "loading";
    elt.style.display = "";
    try {
        const req = await fetch(url);
        if (!req.ok) {
            throw req.statusText;
        }
        elt.style.display = "none";
        return req
    } catch (err) {
        elt.innerText = err.message || err;
        throw err;
    }
}

var excludedEdges = new Set();

// loadGraph loads and displays the lock graph.
async function loadGraph() {
    // Fetch the graph SVG. Unlike say, <object>, this lets us
    // incorporate it directly into the document so events still
    // propagate through the SVG DOM into the HTML DOM.
    const req = await fetchChecked("graph.svg");
    const svgData = await req.text();

    // Insert the SVG into the document.
    const graphView = document.getElementById("graph-view");
    const container = graphView.querySelector(".container");
    container.innerHTML = svgData;
    const lockGraph = container.querySelector("svg");

    // Start loading cycles.
    loadCycles();

    // Make the lock graph zoomable.
    var zoomer = new Zoomer(graphView, container);
    zoomer.shrinkToFit();

    // Hook into graph edges.
    const edges = lockGraph.querySelectorAll(".edge");
    edges.forEach(function(edge) {
        // Get the edge ID.
        const edgeID = edge.id;

        // Increase the size of the click target by making a second,
        // invisible, larger path element.
        const path = edge.querySelector("path");
        const path2 = path.cloneNode(true);
        path2.style.strokeWidth = "10px";
        path2.style.stroke = "transparent";
        path2.style.cursor = "pointer";
        path2.classList.add("overlay");
        path.insertAdjacentElement("afterend", path2);

        // Handle edge clicks.
        path2.addEventListener("click", function(ev) {
            if (ev.ctrlKey) {
                // Toggle edge exclusion.
                if (excludedEdges.has(edgeID))
                    excludedEdges.delete(edgeID);
                else
                    excludedEdges.add(edgeID);
                loadCycles();
                return;
            }

            // Make the lock graph a thumbnail.
            graphView.className = "full thumb";
            // Un-highlight existing highlights.
            lockGraph.querySelectorAll(".edge.selected").forEach(function(edge) {
                edge.classList.remove("selected");
            });
            // Highlight this edge.
            edge.classList.add("selected");
            // TODO: Highlight edges involved in cycles with this edge.
            showEdge(edgeID);
        });
    });
}

// loadCycles loads the set of nodes and edges involved in cycles and
// highlights them in the lock grpah.
async function loadCycles() {
    // Immediately update excluded edge visualization.
    const graphView = document.getElementById("graph-view");
    for (let elt of graphView.querySelectorAll(".excluded")) {
        elt.classList.remove("excluded");
    }
    for (let edge of excludedEdges.values()) {
        document.getElementById(edge).classList.add("excluded");
    }

    // Build query string from excludedEdges.
    var q = [];
    excludedEdges.forEach(v => q.push(v));
    var qs = "";
    if (q.length > 0) {
        qs = "?exc=" + q.join();
    }

    // Fetch cycle data.
    const req2 = await fetchChecked("cycles.json" + qs);
    const cycles = await req2.json();

    // Highlight cycles.
    for (let elt of graphView.querySelectorAll(".in-cycle")) {
        elt.classList.remove("in-cycle");
    }
    for (let node of cycles.nodes) {
        document.getElementById(node).classList.add("in-cycle");
    }
    for (let edge of cycles.edges) {
        document.getElementById(edge).classList.add("in-cycle");
    }
}

// showEdge updates the UI to show information about a given edge.
async function showEdge(edgeID) {
    const edgeView = document.getElementById("edge-view");
    const container = edgeView.querySelector(".container");
    // Delete current contents.
    while (container.firstChild) {
        container.removeChild(container.firstChild);
    }
    // Fetch edge graph.
    const req = await fetchChecked("stacks/" + edgeID + ".svg");
    const svgData = await req.text();
    // Show edge graph.
    container.innerHTML = svgData;
    new Zoomer(edgeView, container).shrinkToFit();
}

// Zoomer makes drags and wheel events on container pan and zoom elt
// within container.
class Zoomer {
    constructor(container, elt) {
        const self = this;

        self._container = container;
        self._elt = elt;

        // We map elt into a unit coordinate space where the short
        // dimension goes from [-1,1] and the long dimension centers
        // [-1,1] while maintaining the aspect ratio of container. We
        // then project that into container.

        // Compute the container projection.
        self._updateProjection();

        // Compute the initial transformation matrix that just cancels
        // out the initial projection matrix.
        elt.style.transformOrigin = "50% 50%";
        self._setMat(self._invProj);

        // Update the projection if the container is resized.
        if (window.ResizeObserver !== undefined) {
            const resizeObserver = new ResizeObserver(function() {
                self._updateProjection();
                self._setMat(self._mat);
            })
            resizeObserver.observe(container);
        }

        // Handle drags.
        var lastpos;
        function pointermove(ev) {
            if (ev.buttons === 0) {
                dragOff(ev);
                return;
            }
            const sDelta = {x: ev.pageX - lastpos.pageX, y: ev.pageY - lastpos.pageY};
            const eDelta = self._matVec(self._invProj, sDelta);
            lastpos = ev;
            const mat = self._matMul({sx: 1, sy: 1, tx: eDelta.x, ty: eDelta.y}, self._mat);
            self._setMat(mat);
            // Don't capture until we're actually moving so things like
            // "click" still work.
            container.setPointerCapture(ev.pointerId);
            ev.preventDefault();
        }
        function dragOff(ev) {
            container.releasePointerCapture(ev.pointerId);
            container.removeEventListener("pointermove", pointermove);
        }
        container.addEventListener("pointerdown", function(ev) {
            lastpos = ev;
            container.addEventListener("pointermove", pointermove);
            ev.preventDefault();
        });
        container.addEventListener("pointerup", function(ev) {
            dragOff(ev);
            ev.preventDefault();
        });

        // Handle zooms.
        container.addEventListener("wheel", function(ev) {
            const delta = ev.deltaY;
            // rates is the delta required to scale by a factor of 2.
            const rates = [
                500, // WheelEvent.DOM_DELTA_PIXEL
                30,  // WheelEvent.DOM_DELTA_LINE
                0.5, // WheelEvent.DOM_DELTA_PAGE
            ];
            const factor = Math.pow(2, -delta / rates[ev.deltaMode]);
            // Translate event coordinates into container space
            // relative to the transform-origin.
            const crect = container.getBoundingClientRect();
            const cPt = {x: ev.clientX - crect.left - crect.width/2,
                         y: ev.clientY - crect.top - crect.height/2};
            // Translate event coordinates into view space.
            const pt = self._matVec(self._invProj, cPt);
            // Scale about point.
            const mat = self._matMul(self._scaleAt(factor, pt.x, pt.y), self._mat);
            self._setMat(mat);
            ev.preventDefault();
        });
    }

    // shrinkToFit shrinks elt to fit into container.
    shrinkToFit() {
        this._elt.style.transform = "none";
        const eltRect = this._elt.getBoundingClientRect();
        const cRect = this._container.getBoundingClientRect();

        if (eltRect.width < cRect.width && eltRect.height < cRect.height) {
            // It fits. Just scale to 1:1 and center it.
            this._setMat(this._invProj);
            return;
        }

        // Compute the "true" bounds of the view space by reverse
        // projecting the top left of the container space to the view
        // space. One of the axes will be [-1,1], but the other will
        // be larger to maintain the aspect ratio.
        const topLeft = this._matVec(this._invProj, {x: -cRect.width/2, y: -cRect.height/2});
        // Fit the bounds of the element to the container bounds (in view space).
        const scale = 2 * Math.min(-topLeft.x / eltRect.width, -topLeft.y / eltRect.height);
        this._setMat({sx:scale, sy:scale, tx:0, ty:0});
    }

    // We represent transformation matrices as four elements:
    //   sx  0 tx
    //    0 sy ty
    //    0  0  1
    _matMul(a, b) {
        return {sx: a.sx * b.sx, sy: a.sy * b.sy,
                tx: a.sx * b.tx + a.tx, ty: a.sy * b.ty + a.ty};
    }
    _matVec(m, v) {
        return {x: m.sx*v.x + m.tx, y: m.sy*v.y + m.ty};
    }
    _matInv(a) {
        return {sx: 1/a.sx, sy: 1/a.sy,
                tx: -a.tx/a.sx, ty: -a.ty/a.sy};
    }
    // Update the projection matrix based on the container's current
    // dimensions.
    _updateProjection() {
        const cRect = this._container.getBoundingClientRect();
        const scale = Math.min(cRect.width/2, cRect.height/2);
        // The transform origin is already at the center.
        this._proj = {sx: scale, sy: scale, tx: 0, ty: 0};
        // Pre-compute inverse projection since we use it often.
        this._invProj = this._matInv(this._proj);
    }
    _scaleAt(scale, tx, ty) {
        // Finally, transform the origin back to (tx,ty).
        return this._matMul({sx: 1, sy: 1, tx: tx, ty: ty},
                            // Then scale around the origin.
                            this._matMul({sx: scale, sy: scale, tx: 0, ty: 0},
                                         // First, transform (tx,ty) to the origin.
                                         {sx: 1, sy: 1, tx: -tx, ty: -ty}));
    }
    _setMat(m) {
        this._mat = m;
        const view = this._matMul(this._proj, m);
        this._elt.style.transform = "matrix(" + view.sx + ",0,0," + view.sy + "," + view.tx + "," + view.ty + ")";
    }
}
