import $ from "jquery";
import Backbone from "backbone";
import { ENTER_KEY } from "./constants";
import Todos from "./todos";
import TodoRouter from "./router";
import AppView from "./app-view";

(function() {
  let app = {};

  let apiRootUrl = "/api/todo";

  // Create our global collection of **Todos**.
  app.todos = new Todos();
  app.todos.url = apiRootUrl;

  app.TodoRouter = new TodoRouter({ app });
  Backbone.history.start();

  // kick things off by creating the `App`
  new AppView({ app });
})();
