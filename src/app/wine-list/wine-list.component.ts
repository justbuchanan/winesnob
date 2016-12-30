import { Input, Component, OnInit, EventEmitter } from '@angular/core';
import { Wine } from '../wine';
import { WineService } from '../wine.service';
import { Router } from '@angular/router';
import { FuzzyByPipe } from 'ng-pipes';


@Component({
  selector: 'app-wine-list',
  templateUrl: './wine-list.component.html',
  styleUrls: ['./wine-list.component.css'],
  providers: [FuzzyByPipe],
})
export class WineListComponent implements OnInit {

  constructor(
      private router: Router,
      private wineService: WineService,
      ) { }

  ngOnInit() {
      this.wineService.getWines().then(wines => {
        this.wines = wines;
      })
  }

  deleteWine(wineId: string) {
    console.log(wineId);

    // remove wine from local store
    this.wineService.deleteWine(wineId).then(success => {
      for (var i = 0; i < this.wines.length; i++) {
        if (this.wines[i].id == wineId) {
          this.wines.splice(i, 1);
          console.log('deleted index ' + i)
          break;
        }
      }
    })
  }

  @Input()
  query: string;

  wines: Wine[] = [];
}
