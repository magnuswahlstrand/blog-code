# Idempotency
A simple fullstack app using
* **React** ([React-Query](https://react-query.tanstack.com/), [Mantine](https://mantine.dev/) for UI)
* **Go** ([Fiber](https://github.com/gofiber/fiber) for routing, [Qmgo](https://github.com/qiniu/qmgo) db-driver)
* **MongoDB**

![search](search.gif)

A short post about it on [wahlstrand.dev](https://wahlstrand.dev/articles/2022-04-15-react-fiber-mongo/).


### Discussion 
* To check if the request is identical to a previous request, we validate only the request body. We might also want to check the headers, and HTTP verb
* Right now, idempotency key is used
* In a real system, we want to split idempotency per client ID or similar to avoid sharing sensitive information between users. 
