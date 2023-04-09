#!/bin/bash

# Useful kubectl commands for Kubeshark development

# This command outputs all Kubernetes resources using YAML format and pipes it to VS Code
if [ $1 = "view-all-resources" ] ; then
  kubectl get $(kubectl api-resources | awk '{print $1}' | tail -n +2 | tr '\n' ',' | sed s/,\$//) -o yaml |  code -
fi

# This command outputs all Kubernetes resources in "kubeshark" namespace using YAML format and pipes it to VS Code
if [[ $1 = "view-kubeshark-resources" ]] ; then
  kubectl get $(kubectl api-resources | awk '{print $1}' | tail -n +2 | tr '\n' ',' | sed s/,\$//) -n kubeshark -o yaml |  code -
fi
