import { Component, OnInit, EventEmitter } from '@angular/core';
import { Wine } from '../wine';
import { WineService } from '../wine.service';
import { Router } from '@angular/router';
import { FuzzyPipe } from 'ng-pipes';

@Component({
  selector: 'app-wine-list',
  templateUrl: './wine-list.component.html',
  styleUrls: ['./wine-list.component.css'],
  providers: [FuzzyPipe],
})
export class WineListComponent implements OnInit {

  constructor(
      private router: Router,
      private wineService: WineService,
      private fuzzy: FuzzyPipe, // fuzzy search filter tied to search field
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

  query: string;

  wines: Wine[] = [];
}