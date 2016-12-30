import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { MyAuthenticationService } from './authentication.service';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  // styleUrls: ['./app.component.css'],
  providers: [],
})
export class LoginComponent {
  constructor(
    private router: Router,
    public authService: MyAuthenticationService,
    ) {}

  // onLogin() {
  //   console.log('login pressed')
  //   this.authService.login().toPromise()
  //       .then(response => {
  //           // var wine: Wine = response.json() as Wine;
  //           // return wine;
  //           this.router.navigate(['/']);
  //           return 'hi';
  //       });
  // }
}
