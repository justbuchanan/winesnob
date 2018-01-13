FROM justbuchanan/docker-archlinux

RUN pacman -Syu --noconfirm
RUN pacman -S --noconfirm nodejs yarn gcc python python-pip go git python2
RUN pacman -Scc --noconfirm
RUN yarn global add @angular/cli
RUN ng version

ENV GOPATH=/go

# relocate into correct dir in go path
ENV DIR=$GOPATH/src/github.com/justbuchanan/winesnob
RUN mkdir -p $DIR
WORKDIR $DIR

COPY package.json yarn.lock ./
RUN yarn install

COPY wine-list.json ./

COPY backend ./backend
RUN go get -v ./backend/...

# copy frontend files and compile, resulting in a statically-servable "dist" directory
COPY protractor.conf.js tslint.json karma.conf.js angular-cli.json ./
COPY src ./src
RUN ng build --env=prod

RUN go build -o winesnob-backend ./backend

VOLUME "/data"
VOLUME "/etc/cellar-config.json"
EXPOSE 8080
CMD ["./winesnob-backend", "--dbpath", "/data/cellar.sqlite3db", "--config", "/etc/cellar-config.json"]
