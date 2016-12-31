FROM justbuchanan/docker-archlinux

RUN pacman -Syu --noconfirm
RUN pacman -S --noconfirm nodejs npm gcc python python-pip go git
RUN npm install -g angular-cli
RUN ng version

RUN mkdir /winesnob
WORKDIR /winesnob

RUN mkdir -p go/src
ENV GOPATH=/winesnob/go
RUN go get -u github.com/gorilla/mux \
    github.com/gorilla/handlers \
    github.com/gorilla/sessions \
    github.com/jinzhu/gorm \
    github.com/jinzhu/gorm/dialects/sqlite \
    golang.org/x/oauth2 \
    golang.org/x/oauth2/google \
    github.com/renstrom/fuzzysearch/fuzzy

COPY package.json ./
RUN npm install

COPY wine-list.json ./

COPY go/src/backend ./go/src/backend

# copy frontend files and compile, resulting in a statically-servable "dist" directory
COPY protractor.conf.js tslint.json karma.conf.js angular-cli.json ./
COPY src ./src
RUN ng build --env=prod

RUN go build backend

VOLUME "/data"
EXPOSE 8080
CMD ["./backend", "--dbpath", "/data/cellar.sqlite3db"]
