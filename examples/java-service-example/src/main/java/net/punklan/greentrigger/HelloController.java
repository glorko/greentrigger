package net.punklan.graftnode;


import io.micronaut.http.annotation.Controller;
import io.micronaut.http.annotation.Get;

@Controller("/hello")
public class HelloController {

    @Get("/")
    public String sayHello() {
        return "Hello, World!";
    }
}