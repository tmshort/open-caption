# Browser smoke test

Drives the real UI in Chrome: Setup rehearsal, Go Live, the Operator page
(Space to advance), and a phone-sized Audience page. It checks that a rehearsal
never reaches the live audience, that a leading loop starts by itself, and that
captions appear on the audience page. Screenshots are written to the directory
given as the first argument.

It needs Node and Chrome. It is not run by `go test`.

```
# 1. a throwaway data directory that already contains the shows' passwords
mkdir -p /tmp/oc-e2e && printf 'admin_password: adminpw\noperator_password: operatorpw\n' > /tmp/oc-e2e/config.yaml
make build && ./bin/open-caption --data /tmp/oc-e2e --http-addr :18095 &

# 2. create the sample show once through the UI (Setup -> "Create a sample show"), then:
npm install puppeteer-core
OC_BASE=http://localhost:18095 node test/e2e/smoke.js /tmp/oc-e2e-shots
```

Environment: `OC_BASE`, `OC_ADMIN_PW`, `CHROME` (path to the Chrome binary).
