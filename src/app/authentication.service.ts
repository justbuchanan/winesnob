import { Injectable } from '@angular/core';
import { Http, Headers, Response } from '@angular/http';
import { Observable } from 'rxjs/Observable';
import { Router } from '@angular/router';
import 'rxjs/add/operator/map';
// import 'rxjs/add/operator/flatMap';
import 'rxjs/add/operator/mergeMap';

@Injectable()
export class MyAuthenticationService {
    constructor(private http: Http, private router: Router) { }
 
    // login() {
    //     return this.http.get('/oauth2/login')
    //         // .map((response: Response) => {
    //         //     // login successful if there's a jwt token in the response
    //         //     let user = response.json();
    //         //     if (user && user.token) {
    //         //         // store user details and jwt token in local storage to keep user logged in between page refreshes
    //         //         localStorage.setItem('currentUser', JSON.stringify(user));
    //         //     }
    //         //     // TODO: hit login-status endpoint
    //         // });

    //         .flatMap((res: Response) => {
    //             console.log(res);
    //             return this.http.get('/oauth2/login-status').map((res: Response) => res.json());
    //         })
    //         .map((info => {
    //             localStorage.setItem('currentUser', info["email"])
    //             console.log(info)
    //         }))
    // }

    getEmail() {
        // return localStorage.getItem('currentUser');
        return this.http.get('/oauth2/login-status')
            .toPromise()
            .then(response => {
                return response.json()["email"];
            });
    }
 
    // logout() {
    //     // remove user from local storage to log user out
        // return this.http.get('/oauth2/logout')
        //     .toPromise()
        //     .then((response: Response) => {
        //         localStorage.removeItem('currentUser');
        //     })
        //     .then(() => {
        //         this.router.navigate(['login'])
        //     });
    // }
}
