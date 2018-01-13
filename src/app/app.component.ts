import { Component, OnInit } from "@angular/core";
import { Router } from "@angular/router";

import { Wine } from "./wine";
import { AuthenticationService } from "./authentication.service";

@Component({
  selector: "app-root",
  templateUrl: "./app.component.html",
  styleUrls: ["./app.component.css"],
  providers: []
})
export class AppComponent implements OnInit {
  constructor(
    private router: Router,
    private authService: AuthenticationService
  ) {}

  ngOnInit(): void {
    this.authService.getEmail().then(email => {
      this.username = email;
    });
  }

  addWineClicked(event) {
    this.router.navigate(["/create"]);
  }

  title = "Cellar";
  query: string;
  wines: Wine[];
  username: string;
}
