# Security Audit Skill

A Kubeshark MCP skill that teaches AI agents to perform systematic Kubernetes
network security audits using the MITRE ATT&CK framework. It examines DNS
queries, HTTP requests, L4 flows, and protocol-level payloads to detect
compromised workloads, C2 communication, data exfiltration, cryptomining,
lateral movement, and credential theft.

See [SKILL.md](SKILL.md) for the full methodology.

## Demo

The demo below shows a real security audit session against a compromised
`k8s-mule` namespace containing 21 workloads, 6 of which were actively
compromised with C2, cryptomining, secret theft, S3 exfiltration, port
scanning, and Redis reconnaissance.

### Claude Code Session

An animated replay of the Claude Code terminal session running the audit:

<!--
  To view this demo, open this README in a browser or any Markdown previewer
  that renders inline HTML. The animation replays the full audit session
  including MCP tool calls, threat analysis, snapshot creation, and evidence
  export.
-->

<div id="ks-term-root" style="background:#0d1117;border-radius:10px;font-family:'SF Mono','Menlo','Monaco','Courier New',monospace;font-size:13px;line-height:1.6;color:#e6edf3;max-width:820px;margin:0 auto;overflow:hidden;box-shadow:0 8px 32px rgba(0,0,0,0.4);">
  <div style="background:#161b22;padding:10px 16px;display:flex;align-items:center;gap:8px;border-bottom:1px solid #30363d;">
    <div style="width:12px;height:12px;border-radius:50%;background:#ff5f56;"></div>
    <div style="width:12px;height:12px;border-radius:50%;background:#ffbd2e;"></div>
    <div style="width:12px;height:12px;border-radius:50%;background:#27c93f;"></div>
    <div style="color:#8b949e;font-size:12px;margin-left:8px;flex:1;text-align:center;">claude code &mdash; security audit session</div>
  </div>
  <div id="ks-term-body" style="padding:16px 20px;height:440px;overflow:hidden;">
    <div id="ks-term-scroll" style="transition:transform 0.5s ease;"></div>
  </div>
</div>

<script>
(function() {
  var C = {
    prompt: 'color:#7ee787',
    cmd: 'color:#e6edf3',
    tool: 'color:#79c0ff',
    toolName: 'color:#d2a8ff;font-weight:600',
    result: 'color:#8b949e',
    critical: 'color:#f85149;font-weight:600',
    high: 'color:#d29922;font-weight:600',
    finding: 'color:#e6edf3;padding-left:4px',
    sep: 'color:#30363d',
    status: 'color:#58a6ff',
    dim: 'color:#484f58',
    section: 'color:#7ee787;font-weight:600;margin-top:4px'
  };
  function s(style, text) { return '<span style="'+style+'">'+text+'</span>'; }

  var steps = [
    { delay: 1400, lines: [
      s(C.prompt,'$ ')+s(C.cmd,'claude'),
      s(C.dim,'Claude Code v1.0.33 &mdash; Opus 4.6'),
    ]},
    { delay: 2200, lines: [
      s(C.prompt,'&gt; ')+s(C.cmd,'find traces for a compromise in the k8s-mule namespace at around 12pm'),
    ]},
    { delay: 1200, lines: [
      s(C.status,'Loading security-audit skill...'),
      s(C.status,'Loading KFL skill...'),
    ]},
    { delay: 1600, lines: [
      '',
      s(C.section,'Section A: Real-Time Analysis'),
    ]},
    { delay: 1400, lines: [
      s(C.tool,'&#9670; ')+s(C.toolName,'check_kubeshark_status'),
      s(C.result,'  Kubeshark running in namespace \'default\'. Tools available.'),
    ]},
    { delay: 1400, lines: [
      s(C.tool,'&#9670; ')+s(C.toolName,'get_data_boundaries'),
      s(C.result,'  oldest: 1779098100999  newest: 1779104778304'),
      s(C.result,'  Data range: 09:55 &mdash; 11:46 UTC'),
      s(C.result,'  12pm BST (UTC+1) = 11:00 UTC &mdash; within range'),
    ]},
    { delay: 1800, lines: [
      '',
      s(C.dim,'  Running 3 queries in parallel...'),
      s(C.tool,'&#9670; ')+s(C.toolName,'list_api_calls')+' '+s(C.dim,'kfl="dns && src.pod.namespace == \'k8s-mule\'"  limit=200'),
      s(C.tool,'&#9670; ')+s(C.toolName,'list_api_calls')+' '+s(C.dim,'kfl="http && src.pod.namespace == \'k8s-mule\'" limit=200'),
      s(C.tool,'&#9670; ')+s(C.toolName,'list_l4_flows')+' '+s(C.dim,'ns=k8s-mule limit=200'),
    ]},
    { delay: 2200, lines: [
      '',
      s(C.section,'DNS Threat Analysis'),
      s(C.result,'  200 DNS calls returned. Analyzing external domains...'),
      '',
      s(C.critical,'  CRITICAL  c2-callback.attacker-infra.example.com'),
      s(C.finding,'           &larr; update-checker-595b7848c9-z7pxz  (8 queries, A + AAAA)'),
      s(C.critical,'  CRITICAL  pool.minexmr.example.com'),
      s(C.finding,'           &larr; batch-processor-588784bd54-bf7ws  (mining pool)'),
    ]},
    { delay: 2200, lines: [
      '',
      s(C.section,'HTTP Traffic Analysis'),
      s(C.critical,'  CRITICAL  resource-syncer-6b9866fb54-mdjpt &rarr; K8s API (172.20.0.1:443)'),
      s(C.finding,'           GET /api/v1/secrets?limit=500             [200] 1,048,576B'),
      s(C.finding,'           GET /api/v1/pods?limit=500               [200] 139,122B'),
      s(C.finding,'           GET /api/v1/configmaps?limit=500          [200] 149,155B'),
      s(C.finding,'           GET /apis/rbac.../clusterrolebindings     [200] 101,171B'),
      s(C.finding,'           GET /api/v1/namespaces?limit=500          [200] 9,977B'),
      s(C.finding,'           GET /api/v1/services?limit=500            [200] 119,394B'),
    ]},
    { delay: 2400, lines: [
      '',
      s(C.section,'L4 External Flows'),
      s(C.result,'  200 flows returned. 137 external from k8s-mule.'),
      '',
      s(C.critical,'  CRITICAL  backup-agent &rarr; 63+ unique AWS S3/STS IPs :443'),
      s(C.finding,'           67.220.251.181  TX:124,670  RX:508,963  (1.2MB)'),
      s(C.finding,'           52.94.230.12    TX:104,120  RX:425,068  (1.0MB)'),
      s(C.finding,'           52.119.199.198  TX:94,530   RX:385,917  (960KB)'),
      s(C.finding,'           ... 60 more destinations'),
    ]},
    { delay: 1800, lines: [
      '',
      s(C.section,'Creating Snapshot for Deep Analysis'),
      s(C.tool,'&#9670; ')+s(C.toolName,'create_snapshot')+' '+s(C.dim,'name="k8s-mule-compromise-investigation" ns=k8s-mule'),
      s(C.result,'  id: 9b0d4b2b  status: in_progress'),
    ]},
    { delay: 1400, lines: [
      s(C.tool,'&#9670; ')+s(C.toolName,'list_api_calls')+' '+s(C.dim,'kfl="redis && src.pod.namespace == \'k8s-mule\'"'),
      s(C.tool,'&#9670; ')+s(C.toolName,'list_api_calls')+' '+s(C.dim,'kfl="http && path.contains(\'c2\')"'),
    ]},
    { delay: 2000, lines: [
      '',
      s(C.high,'  HIGH      session-manager-677b78dc48-nlb42 &rarr; redis-cache :6379'),
      s(C.finding,'           INFO       &mdash; server fingerprinting'),
      s(C.finding,'           CONFIG GET * &mdash; full config dump (7KB)'),
      s(C.finding,'           KEYS *     &mdash; 111,650 bytes of keys enumerated'),
      s(C.finding,'           CLIENT     &mdash; connection list'),
      s(C.finding,'           DBSIZE     &mdash; database size'),
    ]},
    { delay: 1600, lines: [
      '',
      s(C.tool,'&#9670; ')+s(C.toolName,'get_snapshot')+' '+s(C.dim,'id=9b0d4b2b'),
      s(C.result,'  status: completed  size: 239,690,239 bytes (228MB)'),
      s(C.tool,'&#9670; ')+s(C.toolName,'start_snapshot_dissection')+' '+s(C.dim,'snapshot_id=9b0d4b2b'),
      s(C.result,'  dissection_id: 6bf87b81  active: true  progress: 0%'),
    ]},
    { delay: 1600, lines: [
      '',
      s(C.tool,'&#9670; ')+s(C.toolName,'get_api_call_details')+' '+s(C.dim,'Redis CONFIG'),
      s(C.result,'  process: redis-cli  req_body: "GET"  resp_body: "*"'),
      s(C.tool,'&#9670; ')+s(C.toolName,'get_api_call_details')+' '+s(C.dim,'Redis KEYS'),
      s(C.result,'  capture: ebpf  req_body: "*"  local_bytes: 111,650'),
    ]},
    { delay: 1800, lines: [
      '',
      s(C.section,'Snapshot Workload Inventory'),
      s(C.tool,'&#9670; ')+s(C.toolName,'list_workloads')+' '+s(C.dim,'snapshot_id=9b0d4b2b  ns=k8s-mule'),
      s(C.result,'  21 pods. Key IPs:'),
      s(C.result,'  update-checker    10.0.4.153   (NET-002)'),
      s(C.result,'  resource-syncer   10.0.31.132  (NET-009)'),
      s(C.result,'  network-diag      10.0.17.30   (NET-011)'),
      s(C.result,'  backup-agent      10.0.42.2    (NET-014)'),
      s(C.result,'  batch-processor   10.0.56.215  (NET-016)'),
      s(C.result,'  session-manager   10.0.53.219  (NET-019)'),
    ]},
    { delay: 2000, lines: [
      '',
      s(C.section,'Port Scan Detection'),
      s(C.tool,'&#9670; ')+s(C.toolName,'list_l4_flows')+' '+s(C.dim,'ip=10.0.17.30 (network-diagnostics)'),
      s(C.result,'  100 flows returned'),
      '',
      s(C.high,'  HIGH      network-diagnostics &rarr; 20 IPs x 9 ports'),
      s(C.finding,'           10.244.0.10:8443  10.244.0.11:443   10.244.0.12:8080'),
      s(C.finding,'           10.244.0.13:5432  10.244.0.16:3306  10.244.0.1:6379'),
      s(C.finding,'           10.244.0.1:27017  10.244.0.19:9090  ...'),
      s(C.finding,'           All 0 bytes transferred &mdash; SYN scan pattern'),
    ]},
    { delay: 1800, lines: [
      '',
      s(C.tool,'&#9670; ')+s(C.toolName,'list_l4_flows')+' '+s(C.dim,'ip=10.0.4.153 (update-checker)'),
      s(C.critical,'  TCP &rarr; 146.75.34.132:443  TX:2,707  RX:4,209,120  (4.0MB inbound)'),
      s(C.finding,'  UDP &rarr; kube-dns:53  25,674 queries  (C2 beaconing)'),
    ]},
    { delay: 1600, lines: [
      '',
      s(C.tool,'&#9670; ')+s(C.toolName,'get_snapshot_dissection_status'),
      s(C.status,'  progress: 100%  &mdash; dissection complete'),
    ]},
    { delay: 1200, lines: [
      '',
      s(C.section,'Section B: Snapshot Deep Dive (L7)'),
      s(C.tool,'&#9670; ')+s(C.toolName,'list_api_calls')+' '+s(C.dim,'snapshot dns  limit=500'),
      s(C.tool,'&#9670; ')+s(C.toolName,'list_api_calls')+' '+s(C.dim,'snapshot http limit=500'),
    ]},
    { delay: 2200, lines: [
      '',
      s(C.section,'Snapshot DNS (500 calls)'),
      s(C.result,'  Confirmed: c2-callback.attacker-infra.example.com  (6x)'),
      s(C.critical,'  NEW       stratum.pool-mining.example.com          (4x)'),
      s(C.finding,'           &larr; batch-processor (Stratum mining protocol)'),
      s(C.critical,'  NEW       s3.amazonaws.com                         (2x)'),
      s(C.critical,'  NEW       ec2.us-east-1.amazonaws.com              (2x)'),
      s(C.finding,'           &larr; backup-agent (AWS credential usage)'),
    ]},
    { delay: 2200, lines: [
      '',
      s(C.section,'Snapshot HTTP (245 calls)'),
      s(C.result,'  resource-syncer: 93 calls, 89 suspicious'),
      s(C.finding,'    18x GET /api/v1/configmaps     18x GET /api/v1/services'),
      s(C.finding,'    15x GET /api/v1/pods           13x GET /api/v1/namespaces'),
      s(C.finding,'    13x GET /apis/rbac/clusterrolebindings'),
      s(C.finding,'    12x GET /api/v1/secrets        (1MB per response)'),
      '',
      s(C.result,'  backup-agent: 108 calls, all suspicious'),
      s(C.finding,'    108x GET / &rarr; external IPs  [400/401] credential abuse'),
    ]},
    { delay: 1600, lines: [
      '',
      s(C.section,'Evidence Export'),
      s(C.tool,'&#9670; ')+s(C.toolName,'export_snapshot_pcap')+' '+s(C.dim,'bpf="host 10.0.31.132"'),
      s(C.result,'  resource-syncer-api-enum.pcap  (69 pcap files merged)'),
      s(C.tool,'&#9670; ')+s(C.toolName,'export_snapshot_pcap')+' '+s(C.dim,'bpf="host 10.0.4.153"'),
      s(C.tool,'&#9670; ')+s(C.toolName,'download_file')+' '+s(C.dim,'update-checker-c2.pcap'),
      s(C.result,'  447,520 bytes downloaded'),
    ]},
    { delay: 1400, lines: [
      '',
      s(C.tool,'&#9670; ')+s(C.toolName,'write')+' '+s(C.dim,'security-audit-k8s-mule-2026-05-18.md'),
      s(C.tool,'&#9670; ')+s(C.toolName,'npx md-to-pdf')+' '+s(C.dim,'security-audit-k8s-mule-2026-05-18.md'),
      s(C.status,'  PDF generated successfully'),
    ]},
    { delay: 3500, lines: [
      '',
      s(C.sep,'&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;&mdash;'),
      '',
      s(C.critical,'  RESULT: 6 of 21 workloads compromised'),
      '',
      s(C.finding,'  4x CRITICAL: C2, cryptomining, secret theft, S3 exfiltration'),
      s(C.finding,'  2x HIGH:     port scanning, Redis reconnaissance'),
      '',
      s(C.status,'  Report:   security-audit-k8s-mule-2026-05-18.pdf'),
      s(C.status,'  Snapshot: 9b0d4b2b (228MB, fully dissected)'),
      s(C.status,'  PCAPs:   update-checker-c2.pcap, resource-syncer-api-enum.pcap'),
    ]},
  ];

  var currentStep = 0;
  var scrollEl = document.getElementById('ks-term-scroll');
  var bodyEl = document.getElementById('ks-term-body');

  function addLines(lines) {
    lines.forEach(function(line, i) {
      var div = document.createElement('div');
      div.style.cssText = 'opacity:0;white-space:pre-wrap;word-break:break-word;animation:ks-fi 0.3s '+((i*120)+'ms')+' forwards';
      div.innerHTML = line || '&nbsp;';
      scrollEl.appendChild(div);
    });
    var overflow = scrollEl.scrollHeight - bodyEl.clientHeight;
    if (overflow > 0) {
      scrollEl.style.transform = 'translateY(-' + overflow + 'px)';
    }
  }

  function playStep() {
    if (currentStep >= steps.length) {
      setTimeout(function() {
        currentStep = 0;
        scrollEl.innerHTML = '';
        scrollEl.style.transform = '';
        setTimeout(playStep, 800);
      }, 3000);
      return;
    }
    var step = steps[currentStep];
    addLines(step.lines);
    currentStep++;
    setTimeout(playStep, step.delay + (step.lines.length * 120));
  }

  var st = document.createElement('style');
  st.textContent = '@keyframes ks-fi{to{opacity:1}}';
  document.getElementById('ks-term-root').appendChild(st);

  setTimeout(playStep, 1000);
})();
</script>

### Sample Audit Report

The report generated by the audit above. Includes executive summary, threat
table with MITRE ATT&CK mappings, detailed findings with evidence, attack
chain analysis, and remediation steps:

<div id="ks-report" style="background:#fff;border:1px solid #d0d7de;border-radius:8px;max-width:820px;margin:0 auto;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif;font-size:14px;line-height:1.7;color:#1f2328;overflow:hidden;">

  <!-- Header bar -->
  <div style="background:#f6f8fa;padding:12px 20px;border-bottom:1px solid #d0d7de;display:flex;align-items:center;justify-content:space-between;">
    <div style="display:flex;align-items:center;gap:8px;">
      <svg width="16" height="16" viewBox="0 0 16 16" fill="#656d76"><path d="M3.75 1.5a.25.25 0 0 0-.25.25v12.5c0 .138.112.25.25.25h8.5a.25.25 0 0 0 .25-.25V4.664a.25.25 0 0 0-.073-.177l-2.914-2.914a.25.25 0 0 0-.177-.073H3.75zM2 1.75C2 .784 2.784 0 3.75 0h5.586c.464 0 .909.184 1.237.513l2.914 2.914c.329.328.513.773.513 1.237v9.586A1.75 1.75 0 0 1 12.25 16h-8.5A1.75 1.75 0 0 1 2 14.25V1.75z"/></svg>
      <span style="font-weight:600;font-size:13px;color:#1f2328;">security-audit-k8s-mule-2026-05-18.pdf</span>
    </div>
    <span style="font-size:11px;color:#656d76;">Generated by Claude Code + Kubeshark MCP</span>
  </div>

  <!-- Scrollable report body -->
  <div style="max-height:600px;overflow-y:auto;padding:28px 32px;">

    <!-- Title -->
    <h2 style="margin:0 0 4px 0;font-size:22px;font-weight:700;color:#1f2328;border:none;">Kubernetes Network Security Audit Report</h2>

    <!-- Meta -->
    <div style="font-size:13px;color:#656d76;margin-bottom:20px;line-height:1.8;">
      <strong>Cluster:</strong> AWS EKS (us-east-1) &nbsp;&bull;&nbsp;
      <strong>Namespace:</strong> k8s-mule &nbsp;&bull;&nbsp;
      <strong>Date:</strong> 2026-05-18 12:00 BST<br>
      <strong>Audit window:</strong> 10:55 &mdash; 12:46 BST (09:55 &mdash; 11:46 UTC, ~1h 51m)<br>
      <strong>Snapshot:</strong> <code style="background:#f6f8fa;padding:1px 5px;border-radius:3px;font-size:12px;">9b0d4b2b</code> (228MB, full window)
    </div>

    <hr style="border:none;border-top:1px solid #d0d7de;margin:16px 0;">

    <!-- Executive Summary -->
    <h3 style="font-size:16px;font-weight:600;margin:20px 0 8px 0;color:#1f2328;">Executive Summary</h3>
    <p style="margin:0 0 16px 0;">The <code style="background:#f6f8fa;padding:1px 5px;border-radius:3px;font-size:12px;">k8s-mule</code> namespace is <strong style="color:#cf222e;">actively compromised</strong> with a coordinated, multi-stage attack involving 6 of 21 workloads. The attack chain spans the full MITRE ATT&CK kill chain: C2 communication, cryptomining, systematic K8s API secret enumeration (1MB+ of secrets exfiltrated), data exfiltration to 63+ AWS S3 endpoints, internal port scanning across 20 IPs and 9 service ports, and Redis server reconnaissance.</p>

    <!-- Threat Summary Table -->
    <h3 style="font-size:16px;font-weight:600;margin:20px 0 8px 0;color:#1f2328;">Threat Summary</h3>
    <table style="width:100%;border-collapse:collapse;font-size:13px;margin-bottom:16px;">
      <thead>
        <tr style="border-bottom:2px solid #d0d7de;text-align:left;">
          <th style="padding:8px 10px;font-weight:600;">#</th>
          <th style="padding:8px 10px;font-weight:600;">Severity</th>
          <th style="padding:8px 10px;font-weight:600;">Workload</th>
          <th style="padding:8px 10px;font-weight:600;">Threat</th>
          <th style="padding:8px 10px;font-weight:600;">MITRE ATT&CK</th>
        </tr>
      </thead>
      <tbody>
        <tr style="border-bottom:1px solid #eaeef2;">
          <td style="padding:6px 10px;">1</td>
          <td style="padding:6px 10px;"><span style="background:#cf222e;color:#fff;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:600;">CRITICAL</span></td>
          <td style="padding:6px 10px;font-family:monospace;font-size:12px;">update-checker</td>
          <td style="padding:6px 10px;">C2 Command &amp; Control</td>
          <td style="padding:6px 10px;font-size:12px;">T1071.001, T1071.004</td>
        </tr>
        <tr style="border-bottom:1px solid #eaeef2;background:#f6f8fa;">
          <td style="padding:6px 10px;">2</td>
          <td style="padding:6px 10px;"><span style="background:#cf222e;color:#fff;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:600;">CRITICAL</span></td>
          <td style="padding:6px 10px;font-family:monospace;font-size:12px;">batch-processor</td>
          <td style="padding:6px 10px;">Cryptomining</td>
          <td style="padding:6px 10px;font-size:12px;">T1496</td>
        </tr>
        <tr style="border-bottom:1px solid #eaeef2;">
          <td style="padding:6px 10px;">3</td>
          <td style="padding:6px 10px;"><span style="background:#cf222e;color:#fff;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:600;">CRITICAL</span></td>
          <td style="padding:6px 10px;font-family:monospace;font-size:12px;">resource-syncer</td>
          <td style="padding:6px 10px;">K8s API Secret Theft</td>
          <td style="padding:6px 10px;font-size:12px;">T1552.007, T1087.004</td>
        </tr>
        <tr style="border-bottom:1px solid #eaeef2;background:#f6f8fa;">
          <td style="padding:6px 10px;">4</td>
          <td style="padding:6px 10px;"><span style="background:#cf222e;color:#fff;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:600;">CRITICAL</span></td>
          <td style="padding:6px 10px;font-family:monospace;font-size:12px;">backup-agent</td>
          <td style="padding:6px 10px;">Data Exfiltration to AWS S3</td>
          <td style="padding:6px 10px;font-size:12px;">T1537, T1567.002</td>
        </tr>
        <tr style="border-bottom:1px solid #eaeef2;">
          <td style="padding:6px 10px;">5</td>
          <td style="padding:6px 10px;"><span style="background:#9a6700;color:#fff;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:600;">HIGH</span></td>
          <td style="padding:6px 10px;font-family:monospace;font-size:12px;">network-diagnostics</td>
          <td style="padding:6px 10px;">Internal Port Scanning</td>
          <td style="padding:6px 10px;font-size:12px;">T1046</td>
        </tr>
        <tr>
          <td style="padding:6px 10px;">6</td>
          <td style="padding:6px 10px;"><span style="background:#9a6700;color:#fff;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:600;">HIGH</span></td>
          <td style="padding:6px 10px;font-family:monospace;font-size:12px;">session-manager</td>
          <td style="padding:6px 10px;">Redis Reconnaissance</td>
          <td style="padding:6px 10px;font-size:12px;">T1018, T1082</td>
        </tr>
      </tbody>
    </table>

    <hr style="border:none;border-top:1px solid #d0d7de;margin:16px 0;">

    <!-- Finding 1 -->
    <h3 style="font-size:15px;font-weight:600;margin:20px 0 6px 0;color:#1f2328;">Finding 1: C2 Command &amp; Control <span style="background:#cf222e;color:#fff;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:600;vertical-align:middle;">CRITICAL</span></h3>
    <div style="font-size:12px;color:#656d76;margin-bottom:10px;">
      <strong>Workload:</strong> <code style="background:#f6f8fa;padding:1px 5px;border-radius:3px;">update-checker-595b7848c9-z7pxz</code> (10.0.4.153) &nbsp;&bull;&nbsp;
      <strong>MITRE:</strong> T1071.001, T1071.004
    </div>
    <div style="background:#fff8f8;border-left:3px solid #cf222e;padding:10px 14px;margin-bottom:10px;font-size:13px;">
      <strong>Evidence:</strong><br>
      &bull; DNS beaconing: 8 queries to <code style="font-size:12px;">c2-callback.attacker-infra.example.com</code><br>
      &bull; C2 data channel: TCP to <code style="font-size:12px;">146.75.34.132:443</code> &mdash; 2,707 bytes sent, <strong>4,209,120 bytes received</strong> (4.0MB)<br>
      &bull; 25,674 UDP queries to kube-dns &mdash; consistent with C2 polling<br>
      &bull; PCAP: <code style="font-size:12px;">update-checker-c2.pcap</code> (447KB)
    </div>

    <!-- Finding 2 -->
    <h3 style="font-size:15px;font-weight:600;margin:20px 0 6px 0;color:#1f2328;">Finding 2: Cryptomining <span style="background:#cf222e;color:#fff;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:600;vertical-align:middle;">CRITICAL</span></h3>
    <div style="font-size:12px;color:#656d76;margin-bottom:10px;">
      <strong>Workload:</strong> <code style="background:#f6f8fa;padding:1px 5px;border-radius:3px;">batch-processor-588784bd54-bf7ws</code> (10.0.56.215) &nbsp;&bull;&nbsp;
      <strong>MITRE:</strong> T1496
    </div>
    <div style="background:#fff8f8;border-left:3px solid #cf222e;padding:10px 14px;margin-bottom:10px;font-size:13px;">
      <strong>Evidence:</strong><br>
      &bull; Mining pool DNS: 4 queries to <code style="font-size:12px;">pool.minexmr.example.com</code><br>
      &bull; Stratum protocol: 4 queries to <code style="font-size:12px;">stratum.pool-mining.example.com</code><br>
      &bull; Two distinct pools suggest failover configuration
    </div>

    <!-- Finding 3 -->
    <h3 style="font-size:15px;font-weight:600;margin:20px 0 6px 0;color:#1f2328;">Finding 3: K8s API Secret Theft <span style="background:#cf222e;color:#fff;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:600;vertical-align:middle;">CRITICAL</span></h3>
    <div style="font-size:12px;color:#656d76;margin-bottom:10px;">
      <strong>Workload:</strong> <code style="background:#f6f8fa;padding:1px 5px;border-radius:3px;">resource-syncer-6b9866fb54-mdjpt</code> (10.0.31.132) &nbsp;&bull;&nbsp;
      <strong>MITRE:</strong> T1552.007, T1087.004
    </div>
    <div style="background:#fff8f8;border-left:3px solid #cf222e;padding:10px 14px;margin-bottom:10px;font-size:13px;">
      <strong>Evidence:</strong> 93 HTTP GET requests to K8s API (172.20.0.1:443)<br>
      <table style="width:100%;border-collapse:collapse;font-size:12px;margin-top:8px;">
        <tr style="border-bottom:1px solid #eaeef2;">
          <td style="padding:4px 8px;font-family:monospace;">GET /api/v1/secrets?limit=500</td>
          <td style="padding:4px 8px;">12x</td>
          <td style="padding:4px 8px;"><strong>1,048,576B</strong> each</td>
        </tr>
        <tr style="border-bottom:1px solid #eaeef2;">
          <td style="padding:4px 8px;font-family:monospace;">GET /api/v1/configmaps?limit=500</td>
          <td style="padding:4px 8px;">18x</td>
          <td style="padding:4px 8px;">149,155B</td>
        </tr>
        <tr style="border-bottom:1px solid #eaeef2;">
          <td style="padding:4px 8px;font-family:monospace;">GET /api/v1/pods?limit=500</td>
          <td style="padding:4px 8px;">15x</td>
          <td style="padding:4px 8px;">139,122B</td>
        </tr>
        <tr style="border-bottom:1px solid #eaeef2;">
          <td style="padding:4px 8px;font-family:monospace;">GET /apis/rbac.../clusterrolebindings</td>
          <td style="padding:4px 8px;">13x</td>
          <td style="padding:4px 8px;">101,171B</td>
        </tr>
      </table>
      <div style="margin-top:6px;">Total transferred: <strong>~2.2GB</strong></div>
    </div>

    <!-- Finding 4 -->
    <h3 style="font-size:15px;font-weight:600;margin:20px 0 6px 0;color:#1f2328;">Finding 4: Data Exfiltration to AWS S3 <span style="background:#cf222e;color:#fff;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:600;vertical-align:middle;">CRITICAL</span></h3>
    <div style="font-size:12px;color:#656d76;margin-bottom:10px;">
      <strong>Workload:</strong> <code style="background:#f6f8fa;padding:1px 5px;border-radius:3px;">backup-agent-d74c775bb-nbc2p</code> (10.0.42.2) &nbsp;&bull;&nbsp;
      <strong>MITRE:</strong> T1537, T1567.002
    </div>
    <div style="background:#fff8f8;border-left:3px solid #cf222e;padding:10px 14px;margin-bottom:10px;font-size:13px;">
      <strong>Evidence:</strong><br>
      &bull; 137 external TCP connections to 63+ unique AWS IPs on port 443<br>
      &bull; DNS: <code style="font-size:12px;">s3.amazonaws.com</code>, <code style="font-size:12px;">ec2.us-east-1.amazonaws.com</code><br>
      &bull; 108 HTTP requests returning 400/401 &mdash; expired/stolen credentials<br>
      &bull; Top destination: 67.220.251.181 (1.2MB total)
    </div>

    <!-- Finding 5 -->
    <h3 style="font-size:15px;font-weight:600;margin:20px 0 6px 0;color:#1f2328;">Finding 5: Internal Port Scanning <span style="background:#9a6700;color:#fff;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:600;vertical-align:middle;">HIGH</span></h3>
    <div style="font-size:12px;color:#656d76;margin-bottom:10px;">
      <strong>Workload:</strong> <code style="background:#f6f8fa;padding:1px 5px;border-radius:3px;">network-diagnostics-67bf4c7878-tmjks</code> (10.0.17.30) &nbsp;&bull;&nbsp;
      <strong>MITRE:</strong> T1046
    </div>
    <div style="background:#fffbf0;border-left:3px solid #9a6700;padding:10px 14px;margin-bottom:10px;font-size:13px;">
      <strong>Evidence:</strong><br>
      &bull; 100 TCP flows to 20 unique IPs across 9 ports (80, 443, 3306, 5432, 6379, 8080, 8443, 9090, 27017)<br>
      &bull; Target range: 10.244.0.x (cross-namespace pod CIDR)<br>
      &bull; All flows: 0 bytes &mdash; TCP SYN scan
    </div>

    <!-- Finding 6 -->
    <h3 style="font-size:15px;font-weight:600;margin:20px 0 6px 0;color:#1f2328;">Finding 6: Redis Reconnaissance <span style="background:#9a6700;color:#fff;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:600;vertical-align:middle;">HIGH</span></h3>
    <div style="font-size:12px;color:#656d76;margin-bottom:10px;">
      <strong>Workload:</strong> <code style="background:#f6f8fa;padding:1px 5px;border-radius:3px;">session-manager-677b78dc48-nlb42</code> (10.0.53.219) &nbsp;&bull;&nbsp;
      <strong>MITRE:</strong> T1018, T1082
    </div>
    <div style="background:#fffbf0;border-left:3px solid #9a6700;padding:10px 14px;margin-bottom:10px;font-size:13px;">
      <strong>Evidence:</strong> redis-cli against redis-cache (10.0.1.246:6379)<br>
      &bull; <code style="font-size:12px;">INFO</code> &mdash; server fingerprinting<br>
      &bull; <code style="font-size:12px;">CONFIG GET *</code> &mdash; full config dump (7KB)<br>
      &bull; <code style="font-size:12px;">KEYS *</code> &mdash; <strong>111,650 bytes</strong> of keys<br>
      &bull; <code style="font-size:12px;">CLIENT LIST</code> &mdash; connection enumeration<br>
      &bull; <code style="font-size:12px;">DBSIZE</code> &mdash; capacity assessment
    </div>

    <hr style="border:none;border-top:1px solid #d0d7de;margin:16px 0;">

    <!-- Attack Chain -->
    <h3 style="font-size:16px;font-weight:600;margin:20px 0 8px 0;color:#1f2328;">Attack Chain Analysis</h3>
    <div style="background:#f6f8fa;border:1px solid #d0d7de;border-radius:6px;padding:14px 18px;font-family:monospace;font-size:12px;line-height:1.9;margin-bottom:16px;white-space:pre-wrap;">STAGE 1: COMMAND &amp; CONTROL
  &boxur;&horz; update-checker &rarr; c2-callback.attacker-infra.example.com (4MB received)

STAGE 2: RECONNAISSANCE
  &boxur;&horz; network-diagnostics &rarr; Port scan: 20 IPs &times; 9 ports
  &boxur;&horz; session-manager &rarr; Redis CONFIG/KEYS/CLIENT dump
  &boxur;&horz; resource-syncer &rarr; K8s API: secrets, RBAC, pods, services, namespaces

STAGE 3: CREDENTIAL ACCESS
  &boxur;&horz; resource-syncer &rarr; Harvested 1MB+ of K8s Secrets (12 requests)

STAGE 4: EXFILTRATION
  &boxur;&horz; backup-agent &rarr; 137 connections to 63+ AWS S3 IPs (failing 401)

STAGE 5: MONETIZATION
  &boxur;&horz; batch-processor &rarr; Cryptomining via minexmr + stratum pool</div>

    <hr style="border:none;border-top:1px solid #d0d7de;margin:16px 0;">

    <!-- Immediate Actions -->
    <h3 style="font-size:16px;font-weight:600;margin:20px 0 8px 0;color:#1f2328;">Immediate Actions</h3>
    <ol style="margin:0;padding-left:20px;font-size:13px;">
      <li style="margin-bottom:6px;"><strong>Isolate the namespace:</strong> Default-deny NetworkPolicy on k8s-mule (ingress + egress)</li>
      <li style="margin-bottom:6px;"><strong>Kill compromised pods:</strong> Delete all 6 pods</li>
      <li style="margin-bottom:6px;"><strong>Rotate all secrets cluster-wide:</strong> K8s Secrets harvested (1MB+ &times; 12 requests)</li>
      <li style="margin-bottom:6px;"><strong>Revoke AWS IAM credentials:</strong> IRSA/service account creds for k8s-mule pods</li>
      <li style="margin-bottom:6px;"><strong>Rotate Redis session tokens:</strong> All keys enumerated</li>
      <li style="margin-bottom:6px;"><strong>Block C2 domains at DNS:</strong> c2-callback.attacker-infra.example.com, pool.minexmr.example.com, stratum.pool-mining.example.com</li>
      <li style="margin-bottom:6px;"><strong>Audit RBAC:</strong> Revoke cluster-admin bindings for resource-syncer's service account</li>
      <li><strong>Scan container images:</strong> All k8s-mule Deployment images for tampering</li>
    </ol>

    <hr style="border:none;border-top:1px solid #d0d7de;margin:16px 0;">

    <!-- Evidence -->
    <h3 style="font-size:16px;font-weight:600;margin:20px 0 8px 0;color:#1f2328;">Evidence Preservation</h3>
    <table style="width:100%;border-collapse:collapse;font-size:13px;">
      <tr style="border-bottom:1px solid #eaeef2;">
        <td style="padding:6px 10px;font-weight:600;">Snapshot</td>
        <td style="padding:6px 10px;font-family:monospace;font-size:12px;">9b0d4b2b (228MB, fully dissected)</td>
      </tr>
      <tr style="border-bottom:1px solid #eaeef2;background:#f6f8fa;">
        <td style="padding:6px 10px;font-weight:600;">Dissection</td>
        <td style="padding:6px 10px;font-family:monospace;font-size:12px;">6bf87b81 (100% complete)</td>
      </tr>
      <tr style="border-bottom:1px solid #eaeef2;">
        <td style="padding:6px 10px;font-weight:600;">PCAP: C2</td>
        <td style="padding:6px 10px;font-family:monospace;font-size:12px;">update-checker-c2.pcap (447KB)</td>
      </tr>
      <tr>
        <td style="padding:6px 10px;font-weight:600;">PCAP: API enum</td>
        <td style="padding:6px 10px;font-family:monospace;font-size:12px;">resource-syncer-api-enum.pcap</td>
      </tr>
    </table>

  </div>
</div>
