import { BrowserModule } from "@angular/platform-browser";
import { NgModule } from "@angular/core";
import { Router, RouterModule, Routes } from "@angular/router";
import { FormsModule } from "@angular/forms";
import {
  RequestOptions,
  XHRBackend,
  Http,
  HttpModule,
  ConnectionBackend
} from "@angular/http";
import { MatCardModule } from "@angular/material/card";
import { MatToolbarModule } from "@angular/material/toolbar";
import { MatButtonModule } from "@angular/material/button";
import { MatIconModule, MatIconRegistry } from "@angular/material/icon";
import { MatInputModule } from "@angular/material/input";
import { MatCheckboxModule } from "@angular/material/checkbox";
import { NgPipesModule } from "ng-pipes";
import { MatDialogModule } from "@angular/material/dialog";
import { FlexLayoutModule } from "@angular/flex-layout";

import { Angular2FontawesomeModule } from "angular2-fontawesome/angular2-fontawesome";

import { AppComponent } from "./app.component";
import { WineComponent } from "./wine/wine.component";
import { WineEditorComponent } from "./wine-editor/wine-editor.component";
import { WineListComponent } from "./wine-list/wine-list.component";
import { WineService } from "./wine.service";

import { ExtendedHttpService } from "./extended-http.service";
import { AuthenticationService } from "./authentication.service";

import { LoginComponent } from "./login.component";

const appRoutes: Routes = [
  { path: "create", component: WineEditorComponent },
  { path: "edit/:id", component: WineEditorComponent },
  { path: "login", component: LoginComponent }
  // { path: '', component: WineListComponent }
];

export function httpFactory(
  xhrBackend: XHRBackend,
  requestOptions: RequestOptions,
  router: Router
) {
  return new ExtendedHttpService(xhrBackend, requestOptions, router);
}

@NgModule({
  declarations: [
    AppComponent,
    WineComponent,
    LoginComponent,
    WineEditorComponent,
    WineListComponent
  ],
  imports: [
    BrowserModule,
    FormsModule,
    HttpModule,
    MatCardModule,
    MatDialogModule,
    MatToolbarModule,
    MatButtonModule,
    MatIconModule,
    MatInputModule,
    MatCheckboxModule,
    RouterModule.forRoot(appRoutes),
    NgPipesModule,
    FlexLayoutModule,
    Angular2FontawesomeModule
  ],
  providers: [
    MatIconRegistry,
    WineService,
    AuthenticationService,
    {
      provide: ExtendedHttpService,
      useFactory: httpFactory,
      deps: [XHRBackend, RequestOptions, Router]
    }
  ],
  bootstrap: [AppComponent]
})
export class AppModule {}
