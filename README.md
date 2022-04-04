# Project for Golang Intermediate Course

<img src="./go-stripe-vue.png" alt="go app with vue enhancements" />

This repo contains my work on the main project of Trevor Sawler's Udemy course, [Building Web Applications with Go - Intermediate Level](https://www.udemy.com/course/building-web-applications-with-go-intermediate-level/learn/lecture/27168806#overview). The original course concentrates on using Go for an ecommerce application, and I recommend it.

I'm posting this archive, however, because I've used the project course to investigate replacing its javascript with progressive enhancement using Vue 3. There's nothing wrong with Trever's javascript in the course, to be clear. He teaches how to code using plain javascript, sticking close to the standard APIs that all modern browsers implement. This is something any developer should know, or learn. 

But frameworks like Vue, React or Angular are great if a project starts getting large, since you can do elaborate things with the DOM in a very clear way. This is hard to do with the standard javascript APIs, since they take a lot more space to do things that are simple with the dynamic templates of these other libraries.

To get this to work, I've developed a Go module that embeds a Vue build and integrates it with a Go web server. [It's up on Github](https://github.com/torenware/vite-go). If you're interested in the mechanics of getting Vue and Go to work together, I have a much simpler demo application in the module's repo.

It's this app, however, where I've pounded on the integration module and strategy, and worked to get bugs out of the framework. On top of the standard parts of the Go-Stripe app as taught in the course, I've enhanced about a third of the UI to use Vue instead of the original Go-based templates and rendering. These are marked with a small icon (<img src="static/images/gopher.svg" alt="gopher icon" width="15px"> <img src="static/images/vue.svg" alt="vue icon" width="15px">) to show which system is doing the work on that page.

The app demonstrates:

* a reuseable form library with validation.
* a reuseable table with pagination.

In all, the Vue pieces look a lot like the parts of the app that still use the original Go approaches, which is part of the point.

