# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: tests/02-collector-registration.spec.ts >> Collector Registration >> should open registration form
- Location: e2e/tests/02-collector-registration.spec.ts:29:3

# Error details

```
Test timeout of 30000ms exceeded while running "beforeEach" hook.
```

```
Tearing down "context" exceeded the test timeout of 30000ms.
```

# Page snapshot

```yaml
- generic [ref=e3]:
  - generic [ref=e4]:
    - generic [ref=e5]:
      - img "pgAnalytics" [ref=e6]
      - heading "pgAnalytics" [level=1] [ref=e7]
      - paragraph [ref=e8]: PostgreSQL Observability Platform
    - generic [ref=e9]:
      - generic [ref=e10]:
        - heading "Monitor in Real-Time" [level=3] [ref=e11]
        - paragraph [ref=e12]: Track PostgreSQL logs, metrics, and alerts in a unified dashboard
      - generic [ref=e13]:
        - heading "Deep Analysis" [level=3] [ref=e14]
        - paragraph [ref=e15]: Find slow queries, errors, and performance issues instantly
      - generic [ref=e16]:
        - heading "Proactive Alerting" [level=3] [ref=e17]
        - paragraph [ref=e18]: Get notified before problems impact your users
  - generic [ref=e20]:
    - heading "Sign In" [level=2] [ref=e21]
    - paragraph [ref=e22]: Welcome back to pgAnalytics
    - generic [ref=e23]:
      - generic [ref=e24]:
        - generic [ref=e25]: Username*
        - textbox "admin" [ref=e27]
      - generic [ref=e28]:
        - generic [ref=e29]: Password*
        - textbox "••••••••" [ref=e31]
      - generic [ref=e32]:
        - checkbox "Remember me" [ref=e33]
        - generic [ref=e34]: Remember me
      - button "Sign In" [ref=e35] [cursor=pointer]
    - generic [ref=e36]:
      - generic [ref=e41]: Or continue with
      - button "SSO Login" [ref=e42] [cursor=pointer]
    - paragraph [ref=e43]:
      - text: Don't have an account?
      - link "Sign up" [ref=e44] [cursor=pointer]:
        - /url: /signup
```