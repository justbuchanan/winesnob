import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { HttpModule } from '@angular/http';
import { MdCardModule } from '@angular2-material/card';
import { MdToolbarModule } from '@angular2-material/toolbar';
import { MdButtonModule } from '@angular2-material/button';
import { MdIconModule, MdIconRegistry } from '@angular2-material/icon';
import { MdInputModule } from '@angular2-material/input';
import { MdCheckboxModule } from '@angular2-material/checkbox';
import { NgPipesModule } from 'ng-pipes';

import { AppComponent } from './app.component';
import { WineComponent } from './wine/wine.component';
import { WineEditorComponent } from './wine-editor/wine-editor.component';
import { WineListComponent } from './wine-list/wine-list.component';
import { WineService } from './wine.service';

const appRoutes: Routes = [
  { path: 'create', component: WineEditorComponent },
  { path: 'edit/:id', component: WineEditorComponent },
  { path: '', component: WineListComponent }
];

@NgModule({
  declarations: [
    AppComponent,
    WineComponent,
    WineEditorComponent,
    WineListComponent
  ],
  imports: [
    BrowserModule,
    FormsModule,
    HttpModule,
    MdCardModule,
    MdToolbarModule,
    MdButtonModule,
    MdIconModule,
    MdInputModule,
    MdCheckboxModule,
    RouterModule.forRoot(appRoutes),
    NgPipesModule,
  ],
  providers: [
    MdIconRegistry,
    WineService
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
