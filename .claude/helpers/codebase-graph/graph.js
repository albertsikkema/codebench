(function() {
  var allData = ALL_DATA;
  var canvas = document.getElementById('canvas');
  var tooltip = document.getElementById('tooltip');
  var legend = document.getElementById('legend');
  var stats = document.getElementById('stats');
  var infoPanel = document.getElementById('info-panel');
  var currentView = null;
  var cy = null;

  var COLORS = {
    file: '#58a6ff', function: '#7ee787', method: '#d2a8ff', class: '#ff7b72',
    struct: '#ffa657', interface: '#79c0ff', type_alias: '#a5d6ff', enum: '#f2cc60',
    model: '#ff9bce', component: '#b392f0', module: '#8b949e'
  };
  var LANG_COLORS = {
    python: '#3572A5', javascript: '#f1e05a', typescript: '#3178c6',
    go: '#00ADD8', rust: '#dea584', c: '#555555', cpp: '#f34b7d'
  };

  function healthColor(score) {
    var s = Math.max(0, Math.min(1, score));
    var r, g, b;
    if (s < 0.5) {
      var t = s * 2;
      r = Math.round(63 + (210 - 63) * t);
      g = Math.round(185 + (153 - 185) * t);
      b = Math.round(80 + (34 - 80) * t);
    } else {
      var t2 = (s - 0.5) * 2;
      r = Math.round(210 + (248 - 210) * t2);
      g = Math.round(153 + (81 - 153) * t2);
      b = Math.round(34 + (73 - 34) * t2);
    }
    return 'rgb(' + r + ',' + g + ',' + b + ')';
  }

  function healthLabel(pct) {
    return pct < 30 ? 'healthy' : pct < 60 ? 'moderate' : pct < 80 ? 'warm' : 'hot';
  }

  function cleanup() {
    if (cy) { cy.destroy(); cy = null; }
    canvas.innerHTML = '';
    infoPanel.style.display = 'none';
    tooltip.style.display = 'none';
    legend.innerHTML = '';
    stats.innerHTML = '';
    document.querySelector('.graph-controls').classList.remove('active');
    document.querySelector('.health-controls').classList.remove('active');
    window.onresize = null;
  }

  // --- GRAPH VIEW (file / symbol) ---
  function renderGraph(viewName) {
    cleanup();
    document.querySelector('.graph-controls').classList.add('active');

    var data = viewName === 'symbol' ? allData.symbolGraph : allData.fileGraph;
    var isFileView = viewName === 'file';
    var colorMode = isFileView ? 'health' : 'language';

    var colorSelect = document.getElementById('color-mode');
    colorSelect.innerHTML = '';
    ['health', 'coupling', 'language'].forEach(function(v) {
      var opt = document.createElement('option');
      opt.value = v; opt.textContent = v.charAt(0).toUpperCase() + v.slice(1);
      if (v === colorMode) opt.selected = true;
      colorSelect.appendChild(opt);
    });

    var nodeSet = new Set(data.nodes.map(function(n) { return n.id; }));
    var elements = [];

    if (isFileView) {
      var dirs = new Set();
      data.nodes.forEach(function(n) {
        var dir = n.file.indexOf('/') >= 0 ? n.file.substring(0, n.file.lastIndexOf('/')) : '.';
        dirs.add(dir);
      });
      dirs.forEach(function(dir) {
        elements.push({ data: { id: 'dir:' + dir, label: dir, kind: 'directory' }, classes: 'compound' });
      });
      data.nodes.forEach(function(n) {
        var dir = n.file.indexOf('/') >= 0 ? n.file.substring(0, n.file.lastIndexOf('/')) : '.';
        elements.push({ data: {
          id: n.id, label: n.label, kind: n.kind, language: n.language,
          refs: n.refs, symbols: n.symbols, file: n.file, parent: 'dir:' + dir,
          health: n.health, inRefs: n.inRefs, outRefs: n.outRefs, lines: n.lines
        }});
      });
    } else {
      var files = new Set(data.nodes.map(function(n) { return n.file; }));
      files.forEach(function(f) {
        elements.push({ data: { id: 'file:' + f, label: f, kind: 'file-group' }, classes: 'compound' });
      });
      data.nodes.forEach(function(n) {
        elements.push({ data: {
          id: n.id, label: n.label, kind: n.kind, language: n.language,
          refs: n.refs, file: n.file, parent: 'file:' + n.file,
          health: n.health, inRefs: n.inRefs, outRefs: n.outRefs, lines: n.lines
        }});
      });
    }
    var maxWeight = 1;
    data.edges.forEach(function(e) {
      if (e.weight > maxWeight) maxWeight = e.weight;
    });
    data.edges.forEach(function(e) {
      if (nodeSet.has(e.source) && nodeSet.has(e.target))
        elements.push({ data: { source: e.source, target: e.target, type: e.type, weight: e.weight || 1 } });
    });

    // Compute max coupling for ring normalization
    var maxCoupling = 1;
    data.nodes.forEach(function(n) {
      var c = (n.inRefs || 0) + (n.outRefs || 0);
      if (c > maxCoupling) maxCoupling = c;
    });

    cy = cytoscape({
      container: canvas,
      elements: elements,
      style: [
        { selector: 'node', style: {
          'label': 'data(label)', 'font-size': 10, 'color': '#c9d1d9',
          'text-valign': 'bottom', 'text-margin-y': 4,
          'width': function(ele) { return Math.max(16, Math.min(60, 16 + (ele.data('refs') || 0) * 3)); },
          'height': function(ele) { return Math.max(16, Math.min(60, 16 + (ele.data('refs') || 0) * 3)); },
          'background-color': function(ele) {
            if (colorMode === 'health' && ele.data('health') !== undefined) return healthColor(ele.data('health'));
            if (colorMode === 'coupling') {
              var c = (ele.data('inRefs') || 0) + (ele.data('outRefs') || 0);
              return healthColor(c / maxCoupling);
            }
            if (isFileView && ele.data('language')) return LANG_COLORS[ele.data('language')] || COLORS[ele.data('kind')] || '#8b949e';
            return COLORS[ele.data('kind')] || '#8b949e';
          },
          'border-width': function(ele) {
            var c = (ele.data('inRefs') || 0) + (ele.data('outRefs') || 0);
            return Math.max(1, Math.round(1 + 4 * c / maxCoupling));
          },
          'border-color': function(ele) {
            var c = (ele.data('inRefs') || 0) + (ele.data('outRefs') || 0);
            var ratio = c / maxCoupling;
            if (ratio > 0.6) return '#f85149';
            if (ratio > 0.3) return '#d29922';
            return '#30363d';
          },
          'text-max-width': 80, 'text-wrap': 'ellipsis',
        }},
        { selector: 'node.compound, node[kind="directory"], node[kind="file-group"]', style: {
          'background-color': '#161b22', 'background-opacity': 0.6,
          'border-color': '#30363d', 'border-width': 1,
          'label': 'data(label)', 'font-size': 11, 'color': '#8b949e',
          'text-valign': 'top', 'text-halign': 'center', 'text-margin-y': -6,
          'text-max-width': 200, 'text-wrap': 'none', 'padding': 12,
        }},
        { selector: 'edge', style: {
          'width': function(ele) { return Math.max(1, Math.min(5, 1 + 4 * (ele.data('weight') || 1) / maxWeight)); },
          'line-color': function(ele) {
            var w = (ele.data('weight') || 1) / maxWeight;
            if (w > 0.6) return '#f85149';
            if (w > 0.3) return '#d29922';
            return '#30363d';
          },
          'target-arrow-color': function(ele) {
            var w = (ele.data('weight') || 1) / maxWeight;
            if (w > 0.6) return '#f85149';
            if (w > 0.3) return '#d29922';
            return '#30363d';
          },
          'target-arrow-shape': 'triangle', 'curve-style': 'bezier', 'arrow-scale': 0.6,
          'opacity': function(ele) { return Math.max(0.3, Math.min(0.9, 0.3 + 0.6 * (ele.data('weight') || 1) / maxWeight)); },
        }},
        { selector: 'node.highlighted', style: { 'border-color': '#f0e68c', 'border-width': 3, 'z-index': 100 }},
        { selector: 'node.neighbor', style: { 'border-color': '#58a6ff', 'border-width': 2 }},
        { selector: 'edge.highlighted', style: { 'line-color': '#58a6ff', 'target-arrow-color': '#58a6ff', 'opacity': 1, 'width': 2, 'z-index': 100 }},
        { selector: 'node.dimmed', style: { 'opacity': 0.15 }},
        { selector: 'edge.dimmed', style: { 'opacity': 0.05 }},
        { selector: 'node.search-match', style: { 'border-color': '#f0e68c', 'border-width': 3, 'z-index': 100 }},
        { selector: 'node.filtered-out', style: { 'display': 'none' }},
      ],
      layout: { name: 'cose-bilkent', animate: false, nodeDimensionsIncludeLabels: true, idealEdgeLength: 120, nodeRepulsion: 8000 },
      wheelSensitivity: 0.3,
    });

    // Stats
    var nonCompound = cy.nodes().filter(function(n) { return !n.isParent(); });
    var statsHTML = '<strong>' + nonCompound.length + '</strong> nodes &middot; <strong>' + cy.edges().length + '</strong> edges';
    if (isFileView) {
      var hot = nonCompound.filter(function(n) { return (n.data('health') || 0) > 0.7; }).length;
      if (hot > 0) statsHTML += ' &middot; <strong style="color:#f85149">' + hot + '</strong> hot';
    }
    stats.innerHTML = statsHTML;

    function updateLegend() {
      var html = '';
      if (colorMode === 'health' || colorMode === 'coupling') {
        var label = colorMode === 'coupling' ? 'Coupling' : 'Health';
        [{s:0,l:'Low'},{s:0.33,l:'Moderate'},{s:0.66,l:'High'},{s:1,l:'Critical'}].forEach(function(o) {
          html += '<div class="item"><div class="dot" style="background:' + healthColor(o.s) + '"></div>' + o.l + '</div>';
        });
        html += '<div style="margin-top:6px;color:#8b949e;font-size:11px">Ring thickness = coupling degree</div>';
        html += '<div style="color:#8b949e;font-size:11px">Edge thickness = cross-references</div>';
      } else {
        var kinds = [];
        var seen = {};
        data.nodes.forEach(function(n) {
          var k = isFileView ? n.language : n.kind;
          if (k && !seen[k]) { seen[k] = true; kinds.push(k); }
        });
        var cm = isFileView ? LANG_COLORS : COLORS;
        kinds.sort().forEach(function(k) { html += '<div class="item"><div class="dot" style="background:' + (cm[k] || '#8b949e') + '"></div>' + k + '</div>'; });
      }
      legend.innerHTML = html;
    }
    updateLegend();

    colorSelect.onchange = function() { colorMode = this.value; cy.style().update(); updateLegend(); };

    // Node click
    cy.on('tap', 'node', function(evt) {
      var node = evt.target;
      if (node.isParent()) return;
      cy.elements().removeClass('highlighted neighbor dimmed');
      cy.elements().not(node).addClass('dimmed');
      node.addClass('highlighted').removeClass('dimmed');
      var edges = node.connectedEdges();
      edges.addClass('highlighted').removeClass('dimmed');
      var neighbors = node.neighborhood('node');
      neighbors.addClass('neighbor').removeClass('dimmed');
      node.ancestors().removeClass('dimmed');
      neighbors.ancestors().removeClass('dimmed');

      var d = node.data();
      var html = '<h3>' + d.label + '</h3>';
      html += '<div class="field"><span class="field-label">Kind:</span> ' + d.kind + '</div>';
      if (d.file) html += '<div class="field"><span class="field-label">File:</span> ' + d.file + '</div>';
      if (d.language) html += '<div class="field"><span class="field-label">Language:</span> ' + d.language + '</div>';
      if (d.lines) html += '<div class="field"><span class="field-label">Lines:</span> ~' + d.lines + '</div>';
      html += '<div class="field"><span class="field-label">References:</span> ' + (d.refs || 0) + ' in / ' + (d.outRefs || 0) + ' out</div>';
      if (d.symbols) html += '<div class="field"><span class="field-label">Symbols:</span> ' + d.symbols + '</div>';
      if (d.health !== undefined) {
        var pct = Math.round(d.health * 100);
        html += '<div class="field"><span class="field-label">Health:</span> <span style="color:' + healthColor(d.health) + '">' + pct + '% (' + healthLabel(pct) + ')</span></div>';
      }
      var incoming = edges.filter(function(e) { return e.target().id() === node.id(); });
      var outgoing = edges.filter(function(e) { return e.source().id() === node.id(); });
      if (incoming.length > 0) {
        html += '<div class="connections"><strong>Imported by (' + incoming.length + '):</strong><ul>';
        incoming.forEach(function(e) { var src = e.source(); html += '<li data-id="' + src.id() + '">' + src.data('label') + '</li>'; });
        html += '</ul></div>';
      }
      if (outgoing.length > 0) {
        html += '<div class="connections"><strong>Imports (' + outgoing.length + '):</strong><ul>';
        outgoing.forEach(function(e) { var tgt = e.target(); html += '<li data-id="' + tgt.id() + '">' + tgt.data('label') + '</li>'; });
        html += '</ul></div>';
      }
      document.getElementById('info-content').innerHTML = html;
      infoPanel.style.display = 'block';
      document.querySelectorAll('#info-panel .connections li').forEach(function(li) {
        li.addEventListener('click', function() {
          var t = cy.getElementById(this.dataset.id);
          if (t.length) { t.emit('tap'); cy.animate({ center: { eles: t }, zoom: cy.zoom() }, { duration: 300 }); }
        });
      });
    });
    cy.on('tap', function(evt) { if (evt.target === cy) { cy.elements().removeClass('highlighted neighbor dimmed'); infoPanel.style.display = 'none'; } });
    document.getElementById('close-info').onclick = function() { infoPanel.style.display = 'none'; cy.elements().removeClass('highlighted neighbor dimmed'); };

    // Search
    var searchTimeout;
    document.getElementById('search').oninput = function() {
      clearTimeout(searchTimeout);
      var q = this.value.toLowerCase().trim();
      searchTimeout = setTimeout(function() {
        cy.nodes().removeClass('search-match dimmed'); cy.edges().removeClass('dimmed');
        if (!q) return;
        var matches = cy.nodes().filter(function(n) { return !n.isParent() && n.data('label').toLowerCase().indexOf(q) >= 0; });
        if (matches.length > 0) { cy.elements().addClass('dimmed'); matches.forEach(function(m) { m.addClass('search-match').removeClass('dimmed'); m.ancestors().removeClass('dimmed'); }); }
      }, 200);
    };
    document.getElementById('min-refs').onchange = function() {
      var mr = parseInt(this.value) || 0;
      cy.nodes().forEach(function(n) { if (!n.isParent()) { if ((n.data('refs') || 0) < mr) n.addClass('filtered-out'); else n.removeClass('filtered-out'); } });
    };
    document.getElementById('layout-select').onchange = function() {
      var name = this.value;
      var opts = { name: name, animate: true, animationDuration: 500, nodeDimensionsIncludeLabels: true };
      if (name === 'cose-bilkent') { opts.idealEdgeLength = 120; opts.nodeRepulsion = 8000; opts.animate = false; }
      if (name === 'dagre') { opts.rankDir = 'TB'; opts.spacingFactor = 1.2; }
      if (name === 'concentric') { opts.concentric = function(n) { return n.data('refs') || 0; }; opts.levelWidth = function() { return 3; }; }
      cy.layout(opts).run();
    };
    document.getElementById('btn-fit').onclick = function() { cy.fit(null, 30); };
    document.getElementById('btn-reset').onclick = function() {
      cy.elements().removeClass('highlighted neighbor dimmed search-match filtered-out');
      document.getElementById('search').value = ''; document.getElementById('min-refs').value = '0';
      infoPanel.style.display = 'none'; cy.fit(null, 30);
    };
  }

  // --- HEALTH VIEW (circle packing) ---
  function renderHealth() {
    cleanup();
    document.querySelector('.health-controls').classList.add('active');

    var treeData = allData.healthTree;
    var width = canvas.clientWidth;
    var height = canvas.clientHeight;

    var root = d3.hierarchy(treeData)
      .sum(function(d) { return d.value || 0; })
      .sort(function(a, b) { return (b.value || 0) - (a.value || 0); });

    var pack = d3.pack().size([width, height]).padding(3);
    pack(root);

    var svg = d3.select(canvas).append('svg')
      .attr('width', width).attr('height', height).style('background', '#0d1117');

    var focus = root;
    var g = svg.append('g');

    var node = g.selectAll('circle').data(root.descendants()).join('circle')
      .attr('fill', function(d) { if (d.children) return d === root ? 'none' : 'rgba(48, 54, 61, 0.4)'; return healthColor(d.data.health || 0); })
      .attr('stroke', function(d) { return d.children ? '#30363d' : 'none'; })
      .attr('stroke-width', function(d) { return d.children ? 1 : 0; })
      .attr('cursor', function(d) { return d.children ? 'pointer' : 'default'; })
      .on('mouseover', function(event, d) {
        if (d.children) return;
        var dd = d.data;
        var pct = Math.round((dd.health || 0) * 100);
        var html = '<h4>' + (dd.file || dd.name) + '</h4>';
        html += '<div class="field"><span class="field-label">Language:</span> ' + (dd.language || '?') + '</div>';
        html += '<div class="field"><span class="field-label">Symbols:</span> ' + (dd.symbols || 0) + '</div>';
        if (dd.lines) html += '<div class="field"><span class="field-label">Lines:</span> ~' + dd.lines + '</div>';
        html += '<div class="field"><span class="field-label">Refs:</span> ' + (dd.inRefs || 0) + ' in / ' + (dd.outRefs || 0) + ' out</div>';
        html += '<div class="field"><span class="field-label">Health:</span> <span style="color:' + healthColor(dd.health || 0) + '">' + pct + '% (' + healthLabel(pct) + ')</span></div>';
        tooltip.innerHTML = html; tooltip.style.display = 'block';
      })
      .on('mousemove', function(event) { tooltip.style.left = (event.clientX + 12) + 'px'; tooltip.style.top = (event.clientY - 10) + 'px'; })
      .on('mouseout', function() { tooltip.style.display = 'none'; })
      .on('click', function(event, d) { if (d.children && focus !== d) { zoom(d); event.stopPropagation(); } });

    var labels = g.selectAll('text').data(root.descendants()).join('text')
      .attr('text-anchor', 'middle').attr('dominant-baseline', 'central')
      .attr('fill', '#c9d1d9').attr('pointer-events', 'none')
      .style('font-size', '10px').text(function(d) { return d.data.name; });

    function zoomTo(v) {
      var k = Math.min(width, height) / v[2];
      node.attr('transform', function(d) { return 'translate(' + ((d.x - v[0]) * k + width / 2) + ',' + ((d.y - v[1]) * k + height / 2) + ')'; })
          .attr('r', function(d) { return d.r * k; });
      labels.attr('transform', function(d) { return 'translate(' + ((d.x - v[0]) * k + width / 2) + ',' + ((d.y - v[1]) * k + height / 2) + ')'; })
        .style('display', function(d) { var r = d.r * k; if (d.parent !== focus && d !== focus) return 'none'; if (r < 12) return 'none'; return 'block'; })
        .style('font-size', function(d) { return Math.max(8, Math.min(14, d.r * k * 0.3)) + 'px'; });
    }

    function zoom(d) {
      focus = d;
      var t = svg.transition().duration(500);
      node.transition(t).attr('fill-opacity', function(n) { return n.ancestors().indexOf(focus) >= 0 ? 1 : 0.15; });
      zoomTo([focus.x, focus.y, focus.r * 2]);
      updateBreadcrumb();
    }

    function updateBreadcrumb() {
      var path = focus.ancestors().reverse();
      var html = '';
      path.forEach(function(d, i) {
        if (i > 0) html += ' / ';
        if (d === focus) html += '<strong>' + d.data.name + '</strong>';
        else html += '<span data-depth="' + i + '">' + d.data.name + '</span>';
      });
      document.getElementById('breadcrumb').innerHTML = html;
      document.querySelectorAll('#breadcrumb span').forEach(function(s) {
        s.addEventListener('click', function() { zoom(path[parseInt(this.dataset.depth)]); });
      });
    }

    zoomTo([root.x, root.y, root.r * 2]);
    updateBreadcrumb();
    svg.on('click', function() { zoom(root); });
    document.getElementById('btn-zoom-out').onclick = function() { if (focus.parent) zoom(focus.parent); else zoom(root); };

    // Legend
    legend.innerHTML = '<div><strong>Health</strong></div>' +
      '<div class="bar" style="background: linear-gradient(to right, #3fb950, #d29922, #f85149);"></div>' +
      '<div class="labels"><span>Healthy</span><span>Hot</span></div>';

    // Stats
    var leaves = root.leaves();
    var avgHealth = leaves.length > 0 ? leaves.reduce(function(s, d) { return s + (d.data.health || 0); }, 0) / leaves.length : 0;
    var hotCount = leaves.filter(function(d) { return (d.data.health || 0) > 0.7; }).length;
    stats.innerHTML = '<strong>' + leaves.length + '</strong> files &middot; avg health <strong>' +
      Math.round(avgHealth * 100) + '%</strong> &middot; <strong style="color:#f85149">' + hotCount + '</strong> hot';

    window.onresize = function() {
      width = canvas.clientWidth; height = canvas.clientHeight;
      svg.attr('width', width).attr('height', height);
      pack.size([width, height]); pack(root);
      zoomTo([focus.x, focus.y, focus.r * 2]);
    };
  }

  // --- VIEW SWITCHER ---
  document.getElementById('view-select').addEventListener('change', function() {
    switchView(this.value);
  });

  function switchView(view) {
    if (view === currentView) return;
    currentView = view;
    if (view === 'health') renderHealth();
    else renderGraph(view);
  }

  switchView('file');
})();
