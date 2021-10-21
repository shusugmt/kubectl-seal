# kubectl-sealer

[![Build Status](https://github.com/shusugmt/kubectl-sealer/actions/workflows/release.yml/badge.svg)](https://github.com/shusugmt/kubectl-sealer/actions)
[![Go Report Card](https://goreportcard.com/badge/shusugmt/kubectl-sealer)](https://goreportcard.com/report/shusugmt/kubectl-sealer)
[![LICENSE](https://img.shields.io/github/license/shusugmt/kubectl-sealer.svg)](https://github.com/shusugmt/kubectl-sealer/blob/main/LICENSE)
[![Releases](https://img.shields.io/github/release-pre/shusugmt/kubectl-sealer.svg)](https://github.com/shusugmt/kubectl-sealer/releases)

kubectl-sealer brings you a [helm-secrets](https://github.com/jkroepke/helm-secrets) or [sops](https://github.com/mozilla/sops) -like editing experience to your SealedSecret yaml files.


## Installation

You can use [krew](https://krew.sigs.k8s.io/) plugin manager.

```
# need to setup custom index manually, since `krew index add` doesn't support specifying branch
git clone https://github.com/shusugmt/kubectl-sealer.git -b krew-index ~/.krew/index/kubectl-sealer

kubectl krew install kubectl-sealer/sealer
kubectl sealer version
```
