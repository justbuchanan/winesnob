import { Component, OnInit } from '@angular/core';
import { Wine } from './wine';
import { Router } from '@angular/router';

import {MyAuthenticationService} from './authentication.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css'],
  providers: [],
})
export class AppComponent implements OnInit {
  constructor(
    private router: Router,
    private authService: MyAuthenticationService,
    ) {}

  ngOnInit(): void {
    this.authService.getEmail().then(email => { this.username = email; })
    // this.authService.login().toPromise().then(resp => {
    //   this.username = this.authService.getEmail();
    // });
    // console.log(this.username);
  }

  logoutClicked(): void {
    // this.authService.logout();
      // this.ro
    // });
  }

  addWineClicked(event) {
    this.router.navigate(['/create'])
  }

  title = 'Cellar';
  query: string;
  wines: Wine[];
  username: string;
}
