// codebase-graph reads the code-index cache and generates an interactive HTML graph.
//
// Usage:
//
//	go run .claude/helpers/codebase-graph/main.go [options]
//	  -view file|symbol   Graph view (default: file)
//	  -output PATH        Output HTML file (default: codebase-graph.html)
//	  -cache PATH         Index cache file (default: .claude/index-cache/index.json)
//	  -min-refs N         Only include nodes with >= N references (default: 0)
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// --- Index data structures (mirror the MCP server's on-disk format) ---

type Symbol struct {
	Name          string   `json:"name"`
	QualifiedName string   `json:"qualified_name"`
	Kind          string   `json:"kind"`
	FilePath      string   `json:"file_path"`
	Line          int      `json:"line"`
	Language      string   `json:"language"`
	Signature     string   `json:"signature,omitempty"`
	Exported      bool     `json:"exported"`
	BaseClasses   []string `json:"base_classes,omitempty"`
}

type CallEdge struct {
	CallerFile  string `json:"caller_file"`
	CallerScope string `json:"caller_scope"`
	CalleeName  string `json:"callee_name"`
	Line        int    `json:"line"`
	CallType    string `json:"call_type"`
}

type ImportEdge struct {
	ImportingFile string `json:"importing_file"`
	ImportedName  string `json:"imported_name"`
	ImportSource  string `json:"import_source"`
	ResolvedFile  string `json:"resolved_file,omitempty"`
}

type FileInfo struct {
	Path     string `json:"path"`
	Language string `json:"language"`
}

type IndexData struct {
	Symbols     []*Symbol            `json:"symbols"`
	Files       map[string]*FileInfo `json:"files"`
	CallEdges   []*CallEdge          `json:"call_edges"`
	ImportEdges []*ImportEdge        `json:"import_edges"`
}

// --- Graph structures ---

type GraphNode struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	File     string `json:"file,omitempty"`
	Kind     string `json:"kind"`
	Language string `json:"language,omitempty"`
	Refs     int    `json:"refs"`
	Symbols  int    `json:"symbols,omitempty"`
}

type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
}

type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
	View  string      `json:"view"`
}

func main() {
	view := flag.String("view", "file", "Graph view: file (dependencies) or symbol (call graph)")
	output := flag.String("output", "codebase-graph.html", "Output HTML file")
	cachePath := flag.String("cache", ".claude/index-cache/index.json", "Index cache file")
	minRefs := flag.Int("min-refs", 0, "Only include nodes with >= N references")
	flag.Parse()

	if *view != "file" && *view != "symbol" {
		log.Fatal("view must be 'file' or 'symbol'")
	}

	// Find project root
	root := findProjectRoot()

	// Resolve cache path
	cache := *cachePath
	if !filepath.IsAbs(cache) {
		cache = filepath.Join(root, cache)
	}

	// Load index
	data, err := loadIndex(cache)
	if err != nil {
		log.Fatalf("Failed to load index from %s: %v", cache, err)
	}

	if len(data.Symbols) == 0 {
		log.Fatal("Index is empty. Make sure the MCP code-index server has run at least once.")
	}

	// Build graph
	var graph GraphData
	switch *view {
	case "symbol":
		graph = buildSymbolGraph(data, root, *minRefs)
	default:
		graph = buildFileGraph(data, root, *minRefs)
	}

	// Resolve output path
	out := *output
	if !filepath.IsAbs(out) {
		out = filepath.Join(root, out)
	}

	// Write HTML
	if err := writeHTML(graph, out); err != nil {
		log.Fatalf("Failed to write HTML: %v", err)
	}

	fmt.Printf("Generated %s view: %s (%d nodes, %d edges)\n", *view, out, len(graph.Nodes), len(graph.Edges))
}

func findProjectRoot() string {
	cwd, _ := os.Getwd()
	check := cwd
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(check, ".git")); err == nil {
			return check
		}
		parent := filepath.Dir(check)
		if parent == check {
			break
		}
		check = parent
	}
	return cwd
}

func loadIndex(path string) (*IndexData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var data IndexData
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func buildFileGraph(data *IndexData, rootDir string, minRefs int) GraphData {
	// Group symbols by file
	fileSymbols := make(map[string][]*Symbol)
	for _, sym := range data.Symbols {
		fileSymbols[sym.FilePath] = append(fileSymbols[sym.FilePath], sym)
	}

	// Count refs per file: number of unique files that call into this file's symbols
	fileRefs := make(map[string]int)
	callersBySymbol := make(map[string]map[string]bool) // callee_name -> set of caller files
	for _, edge := range data.CallEdges {
		if callersBySymbol[edge.CalleeName] == nil {
			callersBySymbol[edge.CalleeName] = make(map[string]bool)
		}
		callersBySymbol[edge.CalleeName][edge.CallerFile] = true
	}
	for _, sym := range data.Symbols {
		callers := callersBySymbol[sym.Name]
		for callerFile := range callers {
			if callerFile != sym.FilePath {
				fileRefs[sym.FilePath]++
			}
		}
	}

	var nodes []GraphNode
	nodeIDs := make(map[string]bool)
	for path, fi := range data.Files {
		rel, _ := filepath.Rel(rootDir, path)
		refs := fileRefs[path]
		if refs < minRefs {
			continue
		}
		nodes = append(nodes, GraphNode{
			ID:       rel,
			Label:    filepath.Base(rel),
			File:     rel,
			Kind:     "file",
			Language: fi.Language,
			Refs:     refs,
			Symbols:  len(fileSymbols[path]),
		})
		nodeIDs[rel] = true
	}

	var edges []GraphEdge
	seen := make(map[string]bool)
	for _, ie := range data.ImportEdges {
		if ie.ResolvedFile == "" {
			continue
		}
		srcRel, _ := filepath.Rel(rootDir, ie.ImportingFile)
		tgtRel, _ := filepath.Rel(rootDir, ie.ResolvedFile)
		if !nodeIDs[srcRel] || !nodeIDs[tgtRel] {
			continue
		}
		key := srcRel + "|" + tgtRel
		if seen[key] {
			continue
		}
		seen[key] = true
		edges = append(edges, GraphEdge{Source: srcRel, Target: tgtRel, Type: "imports"})
	}

	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Refs > nodes[j].Refs })
	return GraphData{Nodes: nodes, Edges: edges, View: "file"}
}

func buildSymbolGraph(data *IndexData, rootDir string, minRefs int) GraphData {
	// Build lookup maps
	byName := make(map[string][]*Symbol)
	for _, sym := range data.Symbols {
		byName[sym.Name] = append(byName[sym.Name], sym)
	}

	// Count refs per symbol
	callersBySymbol := make(map[string]map[string]bool) // qualified_id -> set of caller files
	for _, edge := range data.CallEdges {
		for _, sym := range byName[edge.CalleeName] {
			qid := sym.FilePath + "::" + sym.QualifiedName
			if callersBySymbol[qid] == nil {
				callersBySymbol[qid] = make(map[string]bool)
			}
			if edge.CallerFile != sym.FilePath {
				callersBySymbol[qid][edge.CallerFile] = true
			}
		}
	}

	var nodes []GraphNode
	nodeIDs := make(map[string]bool)
	for _, sym := range data.Symbols {
		if sym.Kind == "variable" || sym.Kind == "constant" {
			continue
		}
		qid := sym.FilePath + "::" + sym.QualifiedName
		refs := len(callersBySymbol[qid])
		if refs < minRefs {
			continue
		}
		rel, _ := filepath.Rel(rootDir, sym.FilePath)
		nodes = append(nodes, GraphNode{
			ID:       qid,
			Label:    sym.Name,
			File:     rel,
			Kind:     sym.Kind,
			Language: sym.Language,
			Refs:     refs,
		})
		nodeIDs[qid] = true
	}

	var edges []GraphEdge
	seen := make(map[string]bool)
	for _, edge := range data.CallEdges {
		if !nodeIDs[edge.CallerScope] {
			continue
		}
		for _, targetSym := range byName[edge.CalleeName] {
			targetQID := targetSym.FilePath + "::" + targetSym.QualifiedName
			if !nodeIDs[targetQID] || edge.CallerScope == targetQID {
				continue
			}
			key := edge.CallerScope + "|" + targetQID
			if seen[key] {
				continue
			}
			seen[key] = true
			edges = append(edges, GraphEdge{Source: edge.CallerScope, Target: targetQID, Type: "calls"})
		}
	}

	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Refs > nodes[j].Refs })
	return GraphData{Nodes: nodes, Edges: edges, View: "symbol"}
}

func writeHTML(graph GraphData, outputFile string) error {
	graphJSON, err := json.Marshal(graph)
	if err != nil {
		return err
	}

	dir := filepath.Dir(outputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var buf strings.Builder
	buf.WriteString(htmlHead)
	buf.WriteString("\n<script>\nconst GRAPH_DATA = ")
	buf.Write(graphJSON)
	buf.WriteString(";\n</script>\n")
	buf.WriteString(htmlBody)

	return os.WriteFile(outputFile, []byte(buf.String()), 0644)
}

const htmlHead = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Codebase Graph</title>
<style>
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0d1117; color: #c9d1d9; height: 100vh; overflow: hidden; display: flex; flex-direction: column; }
  #toolbar { display: flex; gap: 8px; padding: 8px 12px; background: #161b22; border-bottom: 1px solid #30363d; align-items: center; flex-shrink: 0; z-index: 10; flex-wrap: wrap; }
  #toolbar label { font-size: 13px; color: #8b949e; }
  #toolbar input, #toolbar select, #toolbar button { font-size: 13px; padding: 4px 8px; background: #0d1117; color: #c9d1d9; border: 1px solid #30363d; border-radius: 4px; }
  #toolbar button { cursor: pointer; background: #21262d; }
  #toolbar button:hover { background: #30363d; }
  #search { width: 200px; }
  #cy { flex: 1; }
  #info-panel { position: fixed; right: 12px; top: 56px; width: 300px; max-height: calc(100vh - 68px); overflow-y: auto; background: #161b22; border: 1px solid #30363d; border-radius: 8px; padding: 16px; display: none; z-index: 20; font-size: 13px; }
  #info-panel h3 { font-size: 15px; margin-bottom: 8px; color: #58a6ff; word-break: break-all; }
  #info-panel .field { margin-bottom: 6px; }
  #info-panel .field-label { color: #8b949e; }
  #info-panel .connections { margin-top: 10px; }
  #info-panel .connections ul { list-style: none; padding: 0; }
  #info-panel .connections li { margin: 2px 0; cursor: pointer; color: #58a6ff; }
  #info-panel .connections li:hover { text-decoration: underline; }
  #info-panel .close-btn { position: absolute; top: 8px; right: 12px; cursor: pointer; color: #8b949e; font-size: 18px; }
  #info-panel .close-btn:hover { color: #c9d1d9; }
  #legend { position: fixed; left: 12px; bottom: 12px; background: #161b22; border: 1px solid #30363d; border-radius: 8px; padding: 10px 14px; font-size: 12px; z-index: 10; }
  #legend .item { display: flex; align-items: center; gap: 6px; margin: 3px 0; }
  #legend .dot { width: 10px; height: 10px; border-radius: 50%; }
  #stats { position: fixed; left: 12px; top: 56px; background: #161b22; border: 1px solid #30363d; border-radius: 8px; padding: 10px 14px; font-size: 12px; z-index: 10; color: #8b949e; }
</style>
</head>`

const htmlBody = `<body>
<div id="toolbar">
  <input id="search" type="text" placeholder="Search nodes...">
  <label>Layout:</label>
  <select id="layout-select">
    <option value="cose-bilkent">Force (cose-bilkent)</option>
    <option value="dagre">Hierarchical (dagre)</option>
    <option value="circle">Circle</option>
    <option value="concentric">Concentric</option>
    <option value="grid">Grid</option>
  </select>
  <label>Min refs:</label>
  <input id="min-refs" type="number" value="0" min="0" style="width:60px">
  <button id="btn-fit">Fit</button>
  <button id="btn-reset">Reset</button>
</div>
<div id="cy"></div>
<div id="info-panel">
  <span class="close-btn" id="close-info">&times;</span>
  <div id="info-content"></div>
</div>
<div id="stats"></div>
<div id="legend"></div>

<script src="https://unpkg.com/cytoscape@3.30.4/dist/cytoscape.min.js"></script>
<script src="https://unpkg.com/dagre@0.8.5/dist/dagre.min.js"></script>
<script src="https://unpkg.com/cytoscape-dagre@2.5.0/cytoscape-dagre.js"></script>
<script src="https://unpkg.com/layout-base@2.0.1/layout-base.js"></script>
<script src="https://unpkg.com/cose-base@2.2.0/cose-base.js"></script>
<script src="https://unpkg.com/cytoscape-cose-bilkent@4.1.0/cytoscape-cose-bilkent.js"></script>
<script>
(function() {
  const data = GRAPH_DATA;
  const isFileView = data.view === 'file';

  const COLORS = {
    file: '#58a6ff',
    function: '#7ee787',
    method: '#d2a8ff',
    class: '#ff7b72',
    struct: '#ffa657',
    interface: '#79c0ff',
    type_alias: '#a5d6ff',
    enum: '#f2cc60',
    model: '#ff9bce',
    component: '#b392f0',
    module: '#8b949e',
  };

  const LANG_COLORS = {
    python: '#3572A5',
    javascript: '#f1e05a',
    typescript: '#3178c6',
    go: '#00ADD8',
    rust: '#dea584',
    c: '#555555',
    cpp: '#f34b7d',
  };

  // Build cytoscape elements
  const nodeSet = new Set(data.nodes.map(n => n.id));
  const elements = [];

  if (isFileView) {
    // Group by directory
    const dirs = new Set();
    data.nodes.forEach(n => {
      const dir = n.file.includes('/') ? n.file.substring(0, n.file.lastIndexOf('/')) : '.';
      dirs.add(dir);
    });
    dirs.forEach(dir => {
      elements.push({ data: { id: 'dir:' + dir, label: dir, kind: 'directory' }, classes: 'compound' });
    });
    data.nodes.forEach(n => {
      const dir = n.file.includes('/') ? n.file.substring(0, n.file.lastIndexOf('/')) : '.';
      elements.push({
        data: {
          id: n.id, label: n.label, kind: n.kind, language: n.language,
          refs: n.refs, symbols: n.symbols, file: n.file, parent: 'dir:' + dir
        }
      });
    });
  } else {
    // Symbol view — group by file
    const files = new Set(data.nodes.map(n => n.file));
    files.forEach(f => {
      elements.push({ data: { id: 'file:' + f, label: f, kind: 'file-group' }, classes: 'compound' });
    });
    data.nodes.forEach(n => {
      elements.push({
        data: {
          id: n.id, label: n.label, kind: n.kind, language: n.language,
          refs: n.refs, file: n.file, parent: 'file:' + n.file
        }
      });
    });
  }

  data.edges.forEach(e => {
    if (nodeSet.has(e.source) && nodeSet.has(e.target)) {
      elements.push({ data: { source: e.source, target: e.target, type: e.type } });
    }
  });

  const cy = cytoscape({
    container: document.getElementById('cy'),
    elements: elements,
    style: [
      {
        selector: 'node',
        style: {
          'label': 'data(label)',
          'font-size': 10,
          'color': '#c9d1d9',
          'text-valign': 'bottom',
          'text-margin-y': 4,
          'width': function(ele) { return Math.max(16, Math.min(60, 16 + (ele.data('refs') || 0) * 3)); },
          'height': function(ele) { return Math.max(16, Math.min(60, 16 + (ele.data('refs') || 0) * 3)); },
          'background-color': function(ele) {
            if (isFileView && ele.data('language')) return LANG_COLORS[ele.data('language')] || COLORS[ele.data('kind')] || '#8b949e';
            return COLORS[ele.data('kind')] || '#8b949e';
          },
          'border-width': 1,
          'border-color': '#30363d',
          'text-max-width': 80,
          'text-wrap': 'ellipsis',
        }
      },
      {
        selector: 'node.compound, node[kind="directory"], node[kind="file-group"]',
        style: {
          'background-color': '#161b22',
          'background-opacity': 0.6,
          'border-color': '#30363d',
          'border-width': 1,
          'label': 'data(label)',
          'font-size': 11,
          'color': '#8b949e',
          'text-valign': 'top',
          'text-halign': 'center',
          'text-margin-y': -6,
          'padding': 12,
        }
      },
      {
        selector: 'edge',
        style: {
          'width': 1,
          'line-color': '#30363d',
          'target-arrow-color': '#30363d',
          'target-arrow-shape': 'triangle',
          'curve-style': 'bezier',
          'arrow-scale': 0.6,
          'opacity': 0.5,
        }
      },
      {
        selector: 'node.highlighted',
        style: { 'border-color': '#f0e68c', 'border-width': 3, 'z-index': 100 }
      },
      {
        selector: 'node.neighbor',
        style: { 'border-color': '#58a6ff', 'border-width': 2 }
      },
      {
        selector: 'edge.highlighted',
        style: { 'line-color': '#58a6ff', 'target-arrow-color': '#58a6ff', 'opacity': 1, 'width': 2, 'z-index': 100 }
      },
      {
        selector: 'node.dimmed',
        style: { 'opacity': 0.15 }
      },
      {
        selector: 'edge.dimmed',
        style: { 'opacity': 0.05 }
      },
      {
        selector: 'node.search-match',
        style: { 'border-color': '#f0e68c', 'border-width': 3, 'z-index': 100 }
      },
      {
        selector: 'node.filtered-out',
        style: { 'display': 'none' }
      },
    ],
    layout: { name: 'cose-bilkent', animate: false, nodeDimensionsIncludeLabels: true, idealEdgeLength: 120, nodeRepulsion: 8000 },
    wheelSensitivity: 0.3,
  });

  // Stats
  const nonCompound = cy.nodes().filter(n => !n.isParent());
  document.getElementById('stats').innerHTML =
    '<strong>' + nonCompound.length + '</strong> nodes &middot; <strong>' + cy.edges().length + '</strong> edges';

  // Legend
  const kinds = [...new Set(data.nodes.map(n => isFileView ? n.language : n.kind).filter(Boolean))];
  const colorMap = isFileView ? LANG_COLORS : COLORS;
  let legendHTML = '';
  kinds.sort().forEach(k => {
    legendHTML += '<div class="item"><div class="dot" style="background:' + (colorMap[k] || '#8b949e') + '"></div>' + k + '</div>';
  });
  document.getElementById('legend').innerHTML = legendHTML;

  // Click handler
  cy.on('tap', 'node', function(evt) {
    const node = evt.target;
    if (node.isParent()) return;

    cy.elements().removeClass('highlighted neighbor dimmed');
    cy.elements().not(node).addClass('dimmed');

    node.addClass('highlighted').removeClass('dimmed');
    const connectedEdges = node.connectedEdges();
    connectedEdges.addClass('highlighted').removeClass('dimmed');
    const neighbors = node.neighborhood('node');
    neighbors.addClass('neighbor').removeClass('dimmed');
    node.ancestors().removeClass('dimmed');
    neighbors.ancestors().removeClass('dimmed');

    const d = node.data();
    let html = '<h3>' + d.label + '</h3>';
    html += '<div class="field"><span class="field-label">Kind:</span> ' + d.kind + '</div>';
    if (d.file) html += '<div class="field"><span class="field-label">File:</span> ' + d.file + '</div>';
    if (d.language) html += '<div class="field"><span class="field-label">Language:</span> ' + d.language + '</div>';
    html += '<div class="field"><span class="field-label">References:</span> ' + (d.refs || 0) + '</div>';
    if (d.symbols) html += '<div class="field"><span class="field-label">Symbols:</span> ' + d.symbols + '</div>';

    const incoming = connectedEdges.filter(e => e.target().id() === node.id());
    const outgoing = connectedEdges.filter(e => e.source().id() === node.id());

    if (incoming.length > 0) {
      html += '<div class="connections"><strong>Imported by (' + incoming.length + '):</strong><ul>';
      incoming.forEach(e => {
        const src = e.source();
        html += '<li data-id="' + src.id() + '">' + src.data('label') + '</li>';
      });
      html += '</ul></div>';
    }
    if (outgoing.length > 0) {
      html += '<div class="connections"><strong>Imports (' + outgoing.length + '):</strong><ul>';
      outgoing.forEach(e => {
        const tgt = e.target();
        html += '<li data-id="' + tgt.id() + '">' + tgt.data('label') + '</li>';
      });
      html += '</ul></div>';
    }

    document.getElementById('info-content').innerHTML = html;
    document.getElementById('info-panel').style.display = 'block';

    document.querySelectorAll('#info-panel .connections li').forEach(li => {
      li.addEventListener('click', function() {
        const targetNode = cy.getElementById(this.dataset.id);
        if (targetNode.length) {
          targetNode.emit('tap');
          cy.animate({ center: { eles: targetNode }, zoom: cy.zoom() }, { duration: 300 });
        }
      });
    });
  });

  cy.on('tap', function(evt) {
    if (evt.target === cy) {
      cy.elements().removeClass('highlighted neighbor dimmed');
      document.getElementById('info-panel').style.display = 'none';
    }
  });

  document.getElementById('close-info').addEventListener('click', function() {
    document.getElementById('info-panel').style.display = 'none';
    cy.elements().removeClass('highlighted neighbor dimmed');
  });

  // Search
  let searchTimeout;
  document.getElementById('search').addEventListener('input', function() {
    clearTimeout(searchTimeout);
    const query = this.value.toLowerCase().trim();
    searchTimeout = setTimeout(() => {
      cy.nodes().removeClass('search-match dimmed');
      cy.edges().removeClass('dimmed');
      if (!query) return;
      const matches = cy.nodes().filter(n => !n.isParent() && n.data('label').toLowerCase().includes(query));
      if (matches.length > 0) {
        cy.elements().addClass('dimmed');
        matches.forEach(m => {
          m.addClass('search-match').removeClass('dimmed');
          m.ancestors().removeClass('dimmed');
        });
      }
    }, 200);
  });

  // Min refs filter
  document.getElementById('min-refs').addEventListener('change', function() {
    const minRefs = parseInt(this.value) || 0;
    cy.nodes().forEach(n => {
      if (n.isParent()) return;
      if ((n.data('refs') || 0) < minRefs) {
        n.addClass('filtered-out');
      } else {
        n.removeClass('filtered-out');
      }
    });
  });

  // Layout selector
  document.getElementById('layout-select').addEventListener('change', function() {
    const name = this.value;
    const opts = { name: name, animate: true, animationDuration: 500, nodeDimensionsIncludeLabels: true };
    if (name === 'cose-bilkent') { opts.idealEdgeLength = 120; opts.nodeRepulsion = 8000; opts.animate = false; }
    if (name === 'dagre') { opts.rankDir = 'TB'; opts.spacingFactor = 1.2; }
    if (name === 'concentric') { opts.concentric = function(n) { return n.data('refs') || 0; }; opts.levelWidth = function() { return 3; }; }
    cy.layout(opts).run();
  });

  document.getElementById('btn-fit').addEventListener('click', () => cy.fit(null, 30));
  document.getElementById('btn-reset').addEventListener('click', () => {
    cy.elements().removeClass('highlighted neighbor dimmed search-match filtered-out');
    document.getElementById('search').value = '';
    document.getElementById('min-refs').value = '0';
    document.getElementById('info-panel').style.display = 'none';
    cy.fit(null, 30);
  });
})();
</script>
</body>
</html>`
