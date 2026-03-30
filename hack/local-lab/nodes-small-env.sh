#!/usr/bin/env bash

export K9S_NEO_PROFILE="${K9S_NEO_PROFILE:-k9s-neo-nodes-small}"
export K9S_NEO_COLIMA_PROFILE="${K9S_NEO_COLIMA_PROFILE:-k9s-neo}"
export K9S_NEO_MINIKUBE_NODES="${K9S_NEO_MINIKUBE_NODES:-2}"
export K9S_NEO_DELETE_COLIMA_PROFILE="${K9S_NEO_DELETE_COLIMA_PROFILE:-0}"
export K9S_NEO_MANIFEST="${K9S_NEO_MANIFEST:-$(cd "$(dirname "$0")" && pwd)/manifests/neo-bench-nodes-small.yaml}"
