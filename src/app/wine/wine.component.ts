import { Component, Input, Output, EventEmitter } from "@angular/core";
import { Wine } from "../wine";
import { WineService } from "../wine.service";
import { Router } from "@angular/router";

@Component({
  selector: "app-wine",
  templateUrl: "./wine.component.html",
  styleUrls: ["./wine.component.css"]
})
export class WineComponent {
  constructor(private router: Router, private wineService: WineService) {}

  onEditWine() {
    this.router.navigate(["/edit", this.wine.id]);
  }

  // TODO: should this logic go in the list view?
  // TODO: if so, remove wine service
  // TODO: confirm
  // onDeleteWine() {
  //     this.wineService.deleteWine(this.wine.id)
  //     .then(success => {
  //         // TODO: reload wine list?
  //     });
  // }

  @Input() wine: Wine;

  @Output() delete: EventEmitter<String> = new EventEmitter();
}
