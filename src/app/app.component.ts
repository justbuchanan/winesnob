import { Component, OnInit } from '@angular/core';
import { Wine } from './wine';
import { Router } from '@angular/router';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css'],
  providers: [],
})
export class AppComponent implements OnInit {
  constructor(private router: Router) {}

  ngOnInit(): void {
  }

  addWineClicked(event) {
    this.router.navigate(['/create'])
  }

  title = 'Cellar';
  wines: Wine[];
}
