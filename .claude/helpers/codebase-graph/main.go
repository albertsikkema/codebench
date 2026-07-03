// codebase-graph reads the code-index cache and generates an interactive HTML graph.
//
// Usage:
//
//	go run .claude/helpers/codebase-graph/main.go [options]
//	  -output PATH   Output HTML file (default: .claude/helpers/codebase-graph/codebase-graph.html)
//	  -cache PATH    Index cache file (default: .claude/index-cache/index.json)
//
// Generates a single HTML with three switchable views:
//   - File graph:   Force-directed file dependency graph, colored by language or health
//   - Symbol graph: Force-directed symbol call graph, grouped by file
//   - Health map:   D3 circle-packing diagram showing file health (green=healthy, red=hot)
package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed style.css
var embeddedCSS string

//go:embed graph.js
var embeddedJS string

// --- Index data structures (mirror the MCP server's on-disk format) ---

type Symbol struct {
	Name          string   `json:"name"`
	QualifiedName string   `json:"qualified_name"`
	Kind          string   `json:"kind"`
	FilePath      string   `json:"file_path"`
	Line          int      `json:"line"`
	EndLine       int      `json:"end_line"`
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
	ID       string  `json:"id"`
	Label    string  `json:"label"`
	File     string  `json:"file,omitempty"`
	Kind     string  `json:"kind"`
	Language string  `json:"language,omitempty"`
	Refs     int     `json:"refs"`
	Symbols  int     `json:"symbols,omitempty"`
	Health   float64 `json:"health"` // 0.0 (healthy) to 1.0 (hot)
	InRefs   int     `json:"inRefs"`
	OutRefs  int     `json:"outRefs"`
	Lines    int     `json:"lines,omitempty"`
}

type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
	Weight int    `json:"weight,omitempty"` // number of cross-references between the pair
}

type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
	View  string      `json:"view"`
}

// TreeNode is for circle-packing / treemap views.
type TreeNode struct {
	Name     string      `json:"name"`
	Children []*TreeNode `json:"children,omitempty"`
	Value    int         `json:"value,omitempty"`    // leaf size (symbol count)
	Health   float64     `json:"health"`             // 0-1
	Language string      `json:"language,omitempty"` // leaf only
	File     string      `json:"file,omitempty"`     // leaf only (relative path)
	Lines    int         `json:"lines,omitempty"`
	InRefs   int         `json:"inRefs,omitempty"`
	OutRefs  int         `json:"outRefs,omitempty"`
	Symbols  int         `json:"symbols,omitempty"`
}

// AllData holds all three views for embedding in a single HTML file.
type AllData struct {
	FileGraph   GraphData `json:"fileGraph"`
	SymbolGraph GraphData `json:"symbolGraph"`
	HealthTree  *TreeNode `json:"healthTree"`
}

func main() {
	output := flag.String("output", ".claude/helpers/codebase-graph/codebase-graph.html", "Output HTML file")
	cachePath := flag.String("cache", ".claude/index-cache/index.json", "Index cache file")
	flag.Parse()

	root := findProjectRoot()

	cache := *cachePath
	if !filepath.IsAbs(cache) {
		cache = filepath.Join(root, cache)
	}

	data, err := loadIndex(cache)
	if err != nil {
		log.Fatalf("Failed to load index from %s: %v", cache, err)
	}

	if len(data.Symbols) == 0 {
		log.Fatal("Index is empty. Make sure the MCP code-index server has run at least once.")
	}

	out := *output
	if !filepath.IsAbs(out) {
		out = filepath.Join(root, out)
	}

	all := AllData{
		FileGraph:   buildFileGraph(data, root, 0),
		SymbolGraph: buildSymbolGraph(data, root, 0),
		HealthTree:  buildHealthTree(data, root),
	}

	allJSON, err := json.Marshal(all)
	if err != nil {
		log.Fatalf("Failed to marshal data: %v", err)
	}

	if err := writeCombinedHTML(allJSON, out); err != nil {
		log.Fatalf("Failed to write HTML: %v", err)
	}

	fmt.Printf("Generated: %s (%d files, %d symbols, %d edges)\n",
		out, len(all.FileGraph.Nodes), len(all.SymbolGraph.Nodes), len(all.SymbolGraph.Edges))
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

// projectName derives the project name from the git remote origin URL,
// falling back to the directory basename.
func projectName(rootDir string) string {
	cmd := exec.Command("git", "-C", rootDir, "config", "--get", "remote.origin.url")
	out, err := cmd.Output()
	if err == nil {
		url := strings.TrimSpace(string(out))
		// Handle both SSH (git@...:user/repo.git) and HTTPS (.../user/repo.git)
		if idx := strings.LastIndex(url, "/"); idx >= 0 {
			name := url[idx+1:]
			name = strings.TrimSuffix(name, ".git")
			if name != "" {
				return name
			}
		}
		if idx := strings.LastIndex(url, ":"); idx >= 0 {
			name := url[idx+1:]
			name = strings.TrimSuffix(name, ".git")
			if slash := strings.LastIndex(name, "/"); slash >= 0 {
				name = name[slash+1:]
			}
			if name != "" {
				return name
			}
		}
	}
	return filepath.Base(rootDir)
}

// relPath returns a clean relative path for display. Strips any leading ../
// chains so paths are always relative-looking, even when the index was built
// from a different root directory.
func relPath(rootDir, absPath string) string {
	rel, err := filepath.Rel(rootDir, absPath)
	if err != nil {
		rel = absPath
	}
	// Strip leading ../
	for strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		rel = rel[3:]
	}
	if rel == ".." {
		return filepath.Base(absPath)
	}
	if rel == "" {
		return filepath.Base(absPath)
	}
	return rel
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

	// Count inbound refs per file (unique caller files)
	fileInRefs := make(map[string]int)
	callersBySymbol := make(map[string]map[string]bool)
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
				fileInRefs[sym.FilePath]++
			}
		}
	}

	// Count outbound refs per file (unique callee files)
	fileOutRefs := make(map[string]int)
	calleesByFile := make(map[string]map[string]bool)
	for _, edge := range data.CallEdges {
		for _, sym := range data.Symbols {
			if sym.Name == edge.CalleeName && sym.FilePath != edge.CallerFile {
				if calleesByFile[edge.CallerFile] == nil {
					calleesByFile[edge.CallerFile] = make(map[string]bool)
				}
				calleesByFile[edge.CallerFile][sym.FilePath] = true
			}
		}
	}
	for file, callees := range calleesByFile {
		fileOutRefs[file] = len(callees)
	}

	// Estimate lines per file from symbol byte offsets
	fileLines := make(map[string]int)
	for _, sym := range data.Symbols {
		if sym.EndLine > fileLines[sym.FilePath] {
			fileLines[sym.FilePath] = sym.EndLine
		}
	}

	var nodes []GraphNode
	nodeIDs := make(map[string]bool)
	for path, fi := range data.Files {
		rel := relPath(rootDir, path)
		refs := fileInRefs[path]
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
			InRefs:   fileInRefs[path],
			OutRefs:  fileOutRefs[path],
			Lines:    fileLines[path],
		})
		nodeIDs[rel] = true
	}

	// Compute health scores (normalized 0-1)
	computeHealthScores(nodes)

	edges := make([]GraphEdge, 0)
	edgeWeights := make(map[string]int)
	for _, ie := range data.ImportEdges {
		if ie.ResolvedFile == "" {
			continue
		}
		srcRel := relPath(rootDir, ie.ImportingFile)
		tgtRel := relPath(rootDir, ie.ResolvedFile)
		if !nodeIDs[srcRel] || !nodeIDs[tgtRel] {
			continue
		}
		key := srcRel + "|" + tgtRel
		edgeWeights[key]++
	}
	for key, w := range edgeWeights {
		parts := strings.SplitN(key, "|", 2)
		edges = append(edges, GraphEdge{Source: parts[0], Target: parts[1], Type: "imports", Weight: w})
	}

	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Refs > nodes[j].Refs })
	return GraphData{Nodes: nodes, Edges: edges, View: "file"}
}

// computeHealthScores assigns a 0-1 health score using percentile ranking.
// Combines symbol count, coupling, and file size. Uses rank-based scoring
// so results spread across the full green-to-red range.
func computeHealthScores(nodes []GraphNode) {
	if len(nodes) == 0 {
		return
	}

	// Compute raw scores
	rawScores := make([]float64, len(nodes))
	var maxSymbols, maxCoupling, maxLines int
	for i := range nodes {
		if nodes[i].Symbols > maxSymbols {
			maxSymbols = nodes[i].Symbols
		}
		c := nodes[i].InRefs + nodes[i].OutRefs
		if c > maxCoupling {
			maxCoupling = c
		}
		if nodes[i].Lines > maxLines {
			maxLines = nodes[i].Lines
		}
	}
	for i := range nodes {
		var score float64
		if maxSymbols > 0 {
			score += 0.35 * float64(nodes[i].Symbols) / float64(maxSymbols)
		}
		if maxCoupling > 0 {
			score += 0.40 * float64(nodes[i].InRefs+nodes[i].OutRefs) / float64(maxCoupling)
		}
		if maxLines > 0 {
			score += 0.25 * float64(nodes[i].Lines) / float64(maxLines)
		}
		rawScores[i] = score
	}

	// Rank-normalize: convert raw scores to percentile ranks (0-1)
	// so the spread fills the full color range
	type indexedScore struct {
		idx   int
		score float64
	}
	sorted := make([]indexedScore, len(rawScores))
	for i, s := range rawScores {
		sorted[i] = indexedScore{i, s}
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].score < sorted[j].score })

	n := float64(len(sorted))
	for rank, is := range sorted {
		if n <= 1 {
			nodes[is.idx].Health = 0
		} else {
			nodes[is.idx].Health = float64(rank) / (n - 1)
		}
	}
}

func buildHealthTree(data *IndexData, rootDir string) *TreeNode {
	// Reuse file graph logic for health computation
	graph := buildFileGraph(data, rootDir, 0)

	// Build a tree from flat file list, grouped by directory
	root := &TreeNode{Name: projectName(rootDir)}
	dirNodes := map[string]*TreeNode{"": root}

	// Ensure parent dirs exist
	var ensureDir func(string) *TreeNode
	ensureDir = func(dir string) *TreeNode {
		if n, ok := dirNodes[dir]; ok {
			return n
		}
		parent := filepath.Dir(dir)
		if parent == dir || parent == "." {
			parent = ""
		}
		parentNode := ensureDir(parent)
		n := &TreeNode{Name: filepath.Base(dir)}
		parentNode.Children = append(parentNode.Children, n)
		dirNodes[dir] = n
		return n
	}

	for _, node := range graph.Nodes {
		dir := filepath.Dir(node.File)
		if dir == "." {
			dir = ""
		}
		parentNode := ensureDir(dir)
		val := node.Symbols
		if val < 1 {
			val = 1
		}
		parentNode.Children = append(parentNode.Children, &TreeNode{
			Name:     filepath.Base(node.File),
			Value:    val,
			Health:   node.Health,
			Language: node.Language,
			File:     node.File,
			Lines:    node.Lines,
			InRefs:   node.InRefs,
			OutRefs:  node.OutRefs,
			Symbols:  node.Symbols,
		})
	}

	return root
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
	absToRel := make(map[string]string) // absolute qid -> relative qid for display
	for _, sym := range data.Symbols {
		if sym.Kind == "variable" || sym.Kind == "constant" {
			continue
		}
		qid := sym.FilePath + "::" + sym.QualifiedName
		refs := len(callersBySymbol[qid])
		if refs < minRefs {
			continue
		}
		rel := relPath(rootDir, sym.FilePath)
		relQID := rel + "::" + sym.QualifiedName
		absToRel[qid] = relQID
		nodes = append(nodes, GraphNode{
			ID:       relQID,
			Label:    sym.Name,
			File:     rel,
			Kind:     sym.Kind,
			Language: sym.Language,
			Refs:     refs,
		})
		nodeIDs[qid] = true
	}

	edges := make([]GraphEdge, 0)
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
			edges = append(edges, GraphEdge{Source: absToRel[edge.CallerScope], Target: absToRel[targetQID], Type: "calls"})
		}
	}

	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].File != nodes[j].File {
			return nodes[i].File < nodes[j].File
		}
		return nodes[i].ID < nodes[j].ID
	})
	return GraphData{Nodes: nodes, Edges: edges, View: "symbol"}
}

func writeCombinedHTML(allJSON []byte, outputFile string) error {
	dir := filepath.Dir(outputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var buf strings.Builder
	buf.WriteString(htmlHead)
	buf.WriteString("<style>\n")
	buf.WriteString(embeddedCSS)
	buf.WriteString("</style>\n</head>\n")
	buf.WriteString(htmlBody)
	buf.WriteString("<script>\nconst ALL_DATA = ")
	buf.Write(allJSON)
	buf.WriteString(";\n</script>\n")
	buf.WriteString("<script src=\"https://unpkg.com/cytoscape@3.30.4/dist/cytoscape.min.js\"></script>\n")
	buf.WriteString("<script src=\"https://unpkg.com/dagre@0.8.5/dist/dagre.min.js\"></script>\n")
	buf.WriteString("<script src=\"https://unpkg.com/cytoscape-dagre@2.5.0/cytoscape-dagre.js\"></script>\n")
	buf.WriteString("<script src=\"https://unpkg.com/layout-base@2.0.1/layout-base.js\"></script>\n")
	buf.WriteString("<script src=\"https://unpkg.com/cose-base@2.2.0/cose-base.js\"></script>\n")
	buf.WriteString("<script src=\"https://unpkg.com/cytoscape-cose-bilkent@4.1.0/cytoscape-cose-bilkent.js\"></script>\n")
	buf.WriteString("<script src=\"https://unpkg.com/d3@7.9.0/dist/d3.min.js\"></script>\n")
	buf.WriteString("<script>\n")
	buf.WriteString(embeddedJS)
	buf.WriteString("\n</script>\n</body>\n</html>")

	return os.WriteFile(outputFile, []byte(buf.String()), 0644)
}

const htmlHead = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Codebase Graph</title>
`

const htmlBody = `<body>
<div id="toolbar">
  <label>View:</label>
  <select id="view-select">
    <option value="file">File graph</option>
    <option value="symbol">Symbol graph</option>
    <option value="health">Health map</option>
  </select>
  <div class="sep"></div>
  <div class="graph-controls">
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
    <label>Color:</label>
    <select id="color-mode"></select>
    <button id="btn-fit">Fit</button>
    <button id="btn-reset">Reset</button>
  </div>
  <div class="health-controls">
    <div id="breadcrumb"><span id="root-link">root</span></div>
    <button id="btn-zoom-out">Zoom out</button>
  </div>
</div>
<div id="canvas"></div>
<div id="info-panel">
  <span class="close-btn" id="close-info">&times;</span>
  <div id="info-content"></div>
</div>
<div id="tooltip"></div>
<div id="stats"></div>
<div id="legend"></div>
`
