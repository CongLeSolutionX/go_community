// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package viewer implements an HTML-based interactive viewer for lock
// graphs.
package viewer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"lockcheck/cache"
	"lockcheck/lockgraph"

	"github.com/aclements/go-moremath/graph"
	"github.com/aclements/go-moremath/graph/graphout"
)

// Server is an HTTP server that serves the interactive lock graph
// viewer.
type Server struct {
	// Addr is the network address to listen on, as used by
	// net.Listen. If this is "", the server listens on 127.0.0.1
	// at a random port.
	//
	// Server.Start updates Addr to reflect the full address the
	// server is listening on.
	Addr string

	// Graph is the lock graph to serve.
	Graph *lockgraph.Graph

	// StaticPath is the file system path of this package, for
	// locating static resources.
	StaticPath string

	// graphSVG is a cache of graph SVGs for the lock graph and edge
	// graphs.
	graphSVG cache.Cache
}

// Start starts the HTTP viewer server. It returns once the server is
// listening. It sets s.Addr to the full address the server is
// listening on.
func (s *Server) Start() error {
	s.graphSVG.New = s.makeGraphSVG

	// Open listener.
	address := s.Addr
	if address == "" {
		address = "127.0.0.1:"
	}
	sock, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	// Update s.Addr to the real address we're listening on.
	s.Addr = sock.Addr().String()

	// Create the mux.
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.mkStaticHandler("index.html"))
	mux.HandleFunc("/main.js", s.mkStaticHandler("main.js"))
	mux.HandleFunc("/graph.svg", s.handleGraph)
	mux.HandleFunc("/cycles.json", s.handleCycles)
	mux.HandleFunc("/stacks/", s.handleStacks)

	// Start the server
	hs := &http.Server{
		Handler: mux,
	}
	go hs.Serve(sock)
	return nil
}

func (s *Server) mkStaticHandler(path string) http.HandlerFunc {
	path = filepath.Join(s.StaticPath, path)
	return func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, path)
	}
}

func (s *Server) nodeMapID(node int) interface{} {
	return nodeID{node}.String()
}

func (s *Server) edgeMapID(node, edge int) interface{} {
	return edgeID{node, edge}.String()
}

// handleGraph responds with the SVG of the lock graph.
func (s *Server) handleGraph(w http.ResponseWriter, req *http.Request) {
	svg := s.graphSVG.Get(nil)
	if err, ok := svg.(error); ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "image/svg+xml")
	w.Write(svg.([]byte))
}

// handleCycles responds with a the set of nodes and edges involved in
// cycles.
func (s *Server) handleCycles(w http.ResponseWriter, req *http.Request) {
	var g graph.Graph = s.Graph
	nodeMap, edgeMap := s.nodeMapID, s.edgeMapID

	// Delete excluded edges.
	exc := req.FormValue("exc")
	if exc != "" {
		// Eliminate excluded edges.
		parts := strings.Split(exc, ",")
		rmEdges := []graph.Edge{}
		for _, part := range parts {
			edge, err := parseEdgeID(part)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			rmEdges = append(rmEdges, graph.Edge{edge.nodeID, edge.edgeID})
		}
		subgraph := graph.SubgraphRemove(g, nil, rmEdges)
		g = subgraph
		nodeMap = subgraph.NodeMap(nodeMap)
		edgeMap = subgraph.EdgeMap(edgeMap)
	}

	// Collect cycles.
	cNodes, cEdges := lockgraph.Cycles(g)

	// Map to node and edge IDs.
	type Response struct {
		Nodes []string `json:"nodes"`
		Edges []string `json:"edges"`
	}
	var res Response
	for _, nid := range cNodes {
		res.Nodes = append(res.Nodes, nodeMap(nid).(string))
	}
	for _, edge := range cEdges {
		res.Edges = append(res.Edges, edgeMap(edge.Node, edge.Edge).(string))
	}

	// Respond.
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(res)
}

var stacksRe = regexp.MustCompile(`^/stacks/(.+)\.svg$`)

// handleStacks responds with the SVG of the call graph for an edge in
// the lock graph.
func (s *Server) handleStacks(w http.ResponseWriter, req *http.Request) {
	// Parse URL
	m := stacksRe.FindStringSubmatch(req.URL.Path)
	if m == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	key, err := parseEdgeID(m[1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Get graph SVG.
	svg := s.graphSVG.Get(key)
	if err, ok := svg.(error); ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "image/svg+xml")
	w.Write(svg.([]byte))
}

// makeGraphSVG is a cache-populator that constructs SVGs of the lock
// and edge graphs.
func (s *Server) makeGraphSVG(key interface{}) interface{} {
	// Construct dot code.
	var dotCode bytes.Buffer
	switch key := key.(type) {
	case nil:
		// Lock graph. Give nodes and edges IDs we can pull from the SVG.
		nodeAttrs := func(node int) []graphout.DotAttr {
			return []graphout.DotAttr{
				{Name: "id", Val: nodeID{node}.String()},
			}
		}
		edgeAttrs := func(node, edge int) []graphout.DotAttr {
			return []graphout.DotAttr{
				{Name: "id", Val: edgeID{node, edge}.String()},
			}
		}
		graphout.Dot{Label: s.Graph.Label, NodeAttrs: nodeAttrs, EdgeAttrs: edgeAttrs}.Fprint(&dotCode, s.Graph)
	case edgeID:
		// Edge graph.
		edge := s.Graph.Edges[key.nodeID][key.edgeID]
		s.Graph.EdgeToDot(&dotCode, edge)
	}

	// Convert to SVG.
	var svgCode bytes.Buffer
	var stderr strings.Builder
	dot := exec.Command("dot", "-Tsvg")
	dot.Stdin = &dotCode
	dot.Stdout = &svgCode
	dot.Stderr = &stderr
	err := dot.Run()
	if err != nil {
		log.Print(dotCode.String())
		log.Printf("dot failed with %s:\n%s", err, stderr.String())
		return err
	}
	return svgCode.Bytes()
}

type nodeID struct {
	nodeID int
}

func (n nodeID) String() string {
	return fmt.Sprintf("n%d", n.nodeID)
}

type edgeID struct {
	nodeID, edgeID int
}

func (e edgeID) String() string {
	return fmt.Sprintf("e%d-%d", e.nodeID, e.edgeID)
}

var edgeIDRe = regexp.MustCompile(`^e(\d+)-(\d+)$`)

func parseEdgeID(eid string) (edgeID, error) {
	m := edgeIDRe.FindStringSubmatch(eid)
	if m == nil {
		return edgeID{}, fmt.Errorf("malformed edge ID")
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return edgeID{}, err
	}
	e, err := strconv.Atoi(m[2])
	if err != nil {
		return edgeID{}, err
	}
	return edgeID{n, e}, nil
}
