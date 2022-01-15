---
date: 2011-05-16 09:58:57
keywords: [programming, http]
category: "Best practice"
---

# When to prefer PUT over POST HTTP method?

`PUT` should be idempotent. This means that it's harmless to call a `PUT` request many times. On the contrary, calling `POST` requests repeatedly might change data on the server again.

A way to see it is:

* `PUT` = SQL `UPDATE`
* `POST` = SQL `INSERT`
