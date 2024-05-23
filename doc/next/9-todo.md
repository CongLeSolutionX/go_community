<!-- These items need to be completed and moved to an appropriate location in the release notes. -->

<!-- go.dev/issue/60905, CL 559555 -->
TODO: The new `GOARM64` environment variable needs to be documented. This note should be moved to an appropriate location in the release notes.

<!-- These items need to be reviewed, and mentioned in the Go 1.23 release notes if applicable.

accepted proposal https://go.dev/issue/60951 (from https://go.dev/cl/563398, https://go.dev/cl/563455, https://go.dev/cl/564657)
accepted proposal https://go.dev/issue/61347 (from https://go.dev/cl/549376, https://go.dev/cl/549775, https://go.dev/cl/549815, https://go.dev/cl/549995, https://go.dev/cl/550015, https://go.dev/cl/550735, https://go.dev/cl/550736, https://go.dev/cl/551375, https://go.dev/cl/552955, https://go.dev/cl/558016, https://go.dev/cl/558017, https://go.dev/cl/567455, https://go.dev/cl/580955, https://go.dev/cl/581076)
accepted proposal https://go.dev/issue/61405 (from https://go.dev/cl/557835, https://go.dev/cl/584596)
accepted proposal https://go.dev/issue/61447 (from https://go.dev/cl/516355)
accepted proposal https://go.dev/issue/61476 (from https://go.dev/cl/541135)
accepted proposal https://go.dev/issue/61542 (from https://go.dev/cl/512355)
accepted proposal https://go.dev/issue/62039 (from https://go.dev/cl/559799)
accepted proposal https://go.dev/issue/62292 (from https://go.dev/cl/581555)
accepted proposal https://go.dev/issue/63131 (from https://go.dev/cl/578355)
accepted proposal https://go.dev/issue/63393 (from https://go.dev/cl/543335)
accepted proposal https://go.dev/issue/64548 (from https://go.dev/cl/556820)
accepted proposal https://go.dev/issue/64608 (from https://go.dev/cl/557056)
accepted proposal https://go.dev/issue/64962 (from https://go.dev/cl/558695)
accepted proposal https://go.dev/issue/65573 (from https://go.dev/cl/584218, https://go.dev/cl/584300, https://go.dev/cl/584475, https://go.dev/cl/584476)
accepted proposal https://go.dev/issue/65754 (from https://go.dev/cl/572016)
accepted proposal https://go.dev/issue/65880 (from https://go.dev/cl/569882, https://go.dev/cl/570136, https://go.dev/cl/570139)
accepted proposal https://go.dev/issue/66315 (from https://go.dev/cl/580076)
accepted proposal https://go.dev/issue/66343 (from https://go.dev/cl/571995)
accepted proposal https://go.dev/issue/66456 (from https://go.dev/cl/584655)
accepted proposal https://go.dev/issue/67059 (from https://go.dev/cl/587280)
accepted proposal https://go.dev/issue/67061 (from https://go.dev/cl/586656)
accepted proposal https://go.dev/issue/67065 (from https://go.dev/cl/585856)
accepted proposal https://go.dev/issue/67111 (from https://go.dev/cl/584535)
accepted proposal https://go.dev/issue/44251 (from https://go.dev/cl/529816)
CL 473495 has a RELNOTE comment without a suggested text (from RELNOTE comment in https://go.dev/cl/473495)
CL 564035 has a RELNOTE comment without a suggested text (from RELNOTE comment in https://go.dev/cl/564035)
CL 585556 has a RELNOTE comment without a suggested text (from RELNOTE comment in https://go.dev/cl/585556)
-->

<!-- Maybe should be documented? Maybe shouldn't? Someone familiar with the change needs to determine.

CL 359594 ("x/website/_content/ref/mod: document dotless module paths") - resolved go.dev/issue/32819 ("cmd/go: document that module names without dots are reserved") and also mentioned accepted proposal go.dev/issue/37641
CL 570681 ("os: make FindProcess use pidfd on Linux") mentions accepted proposal go.dev/issue/51246 (described as fully implemented in Go 1.22) and NeedsInvestigation continuation issue go.dev/issue/62654.
-->

<!-- Proposals for golang.org/x repos that don't need to be mentioned here but are picked up by relnote todo.
N/A; these are a part of the section below
-->

<!-- Items that don't need to be mentioned in Go 1.23 release notes but are picked up by relnote todo.

CL 458895 - an x/playground fix that mentioned an accepted cmd/go proposal go.dev/issue/40728 in Go 1.16 milestone...
CL 582097 - an x/build CL working on relnote itself; it doesn't need a release note
CL 561935 - crypto CL that used purego tag and mentioned accepted-but-not-implemented proposal https://go.dev/issue/23172 to document purego tag; doesn't need a release note
CL 568340 - fixed a spurious race in time.Ticker.Reset (added via accepted proposal https://go.dev/issue/33184), doesn't seem to need a release note
CL 562619 - x/website CL documented minimum bootstrap version on go.dev, mentioning accepted proposals go.dev/issue/54265 and go.dev/issue/44505; doesn't need a release note
CL 557055 - x/tools CL implemented accepted proposal https://go.dev/issue/46941 for x/tools/go/ssa
CL 564275 - an x/tools CL that updates test data in preparation for accepted proposal https://go.dev/issue/51473; said proposal isn't implemented for Go 1.23 and so it doesn't need a release note
CL 572535 - used "unix" build tag in more places, mentioned accepted proposal https://go.dev/issue/51572; doesn't need a release note
CL 555255 - an x/tools CL implements accepted proposal https://go.dev/issue/53367 for x/tools/go/cfg
CL 585216 - an x/build CL mentions accepted proposal https://go.dev/issue/56001 because it fixed a bug causing downloads not to be produced for that new-to-Go-1.22 port; this isn't relevant to Go 1.23 release notes
CL 481062 - added examples for accepted proposal https://go.dev/issue/56102; doesn't need a release note
CL 497195 - an x/net CL adds one of 4 fields for accepted proposal https://go.dev/issue/57893 in x/net/http2; seemingly not related to net/http and so doesn't need a Go 1.23 release note
CL 463097, CL 568198 - x/net CLs that implemented accepted proposal https://go.dev/issue/57953 for x/net/websocket; no need for rel note
many x/net CLs - work on accepted proposal https://go.dev/issue/58547 to add a QUIC implementation to x/net/quic
CL 514775 - implements a performance optimization for accepted proposal https://go.dev/issue/59488
CL 484995 - x/sys CL implements accepted proposal https://go.dev/issue/59537 to add x/sys/unix API
CL 555597 - optimizes TypeFor (added in accepted proposal https://go.dev/issue/60088) for non-interface types; doesn't seem to need a release note

-->
