import { Component, OnInit, Input } from '@angular/core';
import { Wine } from '../wine';
import { WineService } from '../wine.service';

import { Router, ActivatedRoute, Params } from '@angular/router';
import 'rxjs/add/operator/switchMap';
import { Observable } from 'rxjs/Observable';

@Component({
  selector: 'app-wine-editor',
  templateUrl: './wine-editor.component.html',
  styleUrls: ['./wine-editor.component.css']
})
export class WineEditorComponent implements OnInit {

  constructor(
      private route: ActivatedRoute,
      private router: Router,
      private wineService: WineService
   ) { }

  ngOnInit() {
      // TODO: is there a better way to handle edit vs create?

      const url: Observable<string> = this.route.url.map(segments => segments.join(''));
      url.subscribe(
          value => this.isNewWine = value == 'create',
          error => console.log(error),
          () => console.log('finished')
      )

      this.route.params
          .switchMap((params: Params) => {
              if (params['id'] != null) {
                  return this.wineService.getWine(params['id'])
              } else {
                  return new Promise<Wine>(function (resolve, reject) {
                      resolve(new Wine);
                  })
              }
          })
          .subscribe((wine: Wine) => this.wine = wine);
  }

  // TODO: error handling
  onSubmit() {
      if (this.isNewWine) {
          console.log('submit')
          this.wineService.createWine(this.wine).then(wine => {
              console.log("created wine!" + JSON.stringify(wine))
              // back to home screen
              this.router.navigate(['/'])
          });
      } else {
          // TODO: update wine
      }
  }

  onCancel() {
      this.router.navigate(['/']);
  }

  @Input()
  wine: Wine;

  isNewWine: boolean;
}
