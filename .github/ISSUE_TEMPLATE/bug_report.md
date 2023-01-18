---
name: Bug report
about: Create a report to help us improve
title: ''
labels: ''
assignees: ''

---

**Describe the bug**
A clear and concise description of what the bug is.

**Provide more information**
Running on EKS, AKS, GKE, Minikube, Rancher, OpenShift? Number of Nodes? CNI?

**To Reproduce**
Steps to reproduce the behavior:
1. Run `kubeshark <command> ...`
2. Click on '...'
3. Scroll down to '...'
4. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Logs**
Upload logs:
1. Run the kubeshark command with `--set dump-logs=true` (e.g `kubeshark tap --set dump-logs=true`)
2. Try to reproduce the issue
3. <kbd>CTRL</kbd>+<kbd>C</kbd> on terminal tab which runs `kubeshark`
4. Upload the logs zip file from `~/.kubeshark/kubeshark_logs_**.zip`

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Desktop (please complete the following information):**
 - OS: [e.g. macOS]
 - Web Browser: [e.g. Google Chrome]

**Additional context**
Add any other context about the problem here.
