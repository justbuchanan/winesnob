FROM justbuchanan/docker-archlinux

RUN pacman -Syu --noconfirm
RUN pacman -S --noconfirm nodejs npm gcc python python-pip go git
RUN npm install -g angular-cli
RUN ng version

ENV GOPATH=/go
RUN go get -u github.com/gorilla/mux \
    github.com/gorilla/handlers \
    github.com/gorilla/sessions \
    github.com/jinzhu/gorm \
    github.com/jinzhu/gorm/dialects/sqlite \
    golang.org/x/oauth2 \
    golang.org/x/oauth2/google \
    github.com/renstrom/fuzzysearch/fuzzy

# relocate into correct dir in go path
ENV DIR=$GOPATH/src/github.com/justbuchanan/winesnob
RUN mkdir -p $DIR
WORKDIR $DIR

COPY package.json ./
RUN npm install

COPY wine-list.json ./

COPY backend ./backend

# copy frontend files and compile, resulting in a statically-servable "dist" directory
COPY protractor.conf.js tslint.json karma.conf.js angular-cli.json ./
COPY src ./src
RUN ng build --env=prod

RUN go build -o winesnob-backend ./backend

VOLUME "/data"
VOLUME "/etc/cellar-config.json"
EXPOSE 8080
CMD ["./winesnob-backend", "--dbpath", "/data/cellar.sqlite3db", "--config", "/etc/cellar-config.json"]
