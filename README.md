<!--
SPDX-FileCopyrightText: 2025 Canonical Ltd
SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
Copyright 2019 free5GC.org

SPDX-License-Identifier: Apache-2.0
-->
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/omec-project/udr/badge)](https://scorecard.dev/viewer/?uri=github.com/omec-project/udr)

# UDR

Implements 3gpp 29.504 specification. Provides service to PCF, UDM. UDR supports
SBI interface and any other network function can use the service.

## What this fork adds

This branch wires four route groups that were present in the API surface but answered
`501 Not Implemented`, so nothing sent to them could be stored or read back.

| Route group | Before | Now |
|---|---|---|
| `GET/PUT/DELETE /application-data/pfds` | 501 | served |
| `GET/PUT/PATCH/DELETE /application-data/influenceData` | 501 | served |
| `GET/PUT/DELETE /policy-data/bdt-data` | 501 | served |

The handlers were stubs. Each logged a warning and called
`writeNotImplementedProblem`, while the logic that should have served the request already
existed in `producer/` and was never reached. The change is routing and response
translation, not new business logic: the handlers now call their producers and a shared
helper in `producer_response.go` maps the producer result onto the HTTP response.

Seven files, 181 lines added.

### Why it matters

Two consumers of this data were broken by it, and one failed in a way that pointed at the
wrong component.

**Packet Flow Descriptions could not persist.** An application supplies the rules the core
needs in order to recognise its traffic, domain names and three tuples, through the NEF.
Those rules are stored here. With the route returning 501, nothing an application pushed
survived, and the PFD management service was unusable even though it was implemented.

**Background data transfer failed before any NEF code ran.** The PCF reads
`policy-data/bdt-data` on every BDT create and treats anything other than 200 as fatal. A
501 from this route made the PCF answer the application `503`, so the whole TS 29.554
negotiation appeared to be a PCF problem when the cause was a missing route here.

### Checking it

From inside the cluster, ask the UDR directly:

```bash
kubectl -n aether-5gc run udr-$RANDOM --rm --attach --restart=Never --quiet \
  --image=curlimages/curl:8.7.1 -- sh -c '
for p in application-data/pfds application-data/influenceData policy-data/bdt-data; do
  printf "%-34s %s\n" "$p" "$(curl -sk -o /dev/null -w "%{http_code}" https://udr:29504/nudr-dr/v2/$p)"
done'
```

Expect `200`, `400`, `200`. `influenceData` answers 400 because it requires query
parameters. What matters is that none of the three returns 501, which is the stock answer
and means the route is not mounted.

### Related repositories

- [NEF](https://github.com/TOSSI-Foundation/NEF), the Network Exposure Function that writes
  and reads this data on behalf of applications
- [sdcore-nef-dep](https://github.com/TOSSI-Foundation/sdcore-nef-dep), deployment
  manifests, build scripts and documentation for the whole integration

## UDR flow diagram
![UDR Flow Diagram](/docs/images/README-UDR.png)

## Dynamic Network configuration (via webconsole)

UDR polls the webconsole every 5 seconds to fetch the latest PLMN configuration.

### Setting Up Polling

Include the `webuiUri` of the webconsole in the configuration file
```
configuration:
  ...
  webuiUri: https://webui:5001 # or http://webui:5001
  ...
```
The scheme (http:// or https://) must be explicitly specified. If no parameter is specified,
UDR will use `http://webui:5001` by default.

### HTTPS Support

If the webconsole is served over HTTPS and uses a custom or self-signed certificate,
you must install the root CA certificate into the trust store of the UDR environment.

Check the official guide for installing root CA certificates on Ubuntu:
[Install a Root CA Certificate in the Trust Store](https://documentation.ubuntu.com/server/how-to/security/install-a-root-ca-certificate-in-the-trust-store/index.html)

## Upcoming changes
- Subscription management callbacks to network functions.

Compliance of the 5G Network functions can be found at [5G Compliance](https://docs.sd-core.opennetworking.org/main/overview/3gpp-compliance-5g.html)

## Reach out to us through

1. #sdcore-dev channel in [ONF Community Slack](https://aether5g-project.slack.com)
2. Raise Github [issues](https://github.com/omec-project/udr/issues/new)
