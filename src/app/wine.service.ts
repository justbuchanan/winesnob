import { Injectable } from '@angular/core';
import { Headers } from '@angular/http';
import { ExtendedHttpService } from './extended-http.service';
import 'rxjs/add/operator/toPromise';

import { Wine } from './wine';

@Injectable()
export class WineService {
    constructor(private http: ExtendedHttpService) { }

    getWines(): Promise<Wine[]> {
        return this.http.get('api/wines')
            .toPromise()
            .then(response => {
                var wines: Wine[] = response.json() as Wine[];
                return wines;
            })
            .then(wines => {
                return wines;
            })
            .catch(this.handleError);
    }

    createWine(wine: Wine): Promise<Wine> {
        return this.http.post('api/wines', JSON.stringify(wine))
        .toPromise()
        .then(response => {
            var wine: Wine = response.json() as Wine;
            return wine;
        })
        .catch(this.handleError);
    }

    getWine(id: string): Promise<Wine> {
        return this.http.get('api/wine/' + id)
        .toPromise()
        .then(response => {
            var wine: Wine = response.json() as Wine;
            return wine;
        })
        .catch(this.handleError);
    }

    // TODO: check response code?
    // TODO: Promise return value?
    deleteWine(id: string): Promise<boolean> {
        return this.http.delete('api/wine/' + id)
        .toPromise()
        .then(response => {
            return true;
        })
        .catch(this.handleError);
    }

    private handleError(error: any): Promise<any> {
        // console.error('An error occurred', error);
        return Promise.reject(error.message || error);
    }

}
