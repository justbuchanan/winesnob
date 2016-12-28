FROM justbuchanan/docker-archlinux

RUN pacman -Syu --noconfirm
RUN pacman -S --noconfirm nodejs npm gcc python python-pip go git
# RUN npm install -g angular-cli
# RUN ng version

RUN mkdir /winesnob
WORKDIR /winesnob

RUN mkdir -p go/src
ENV GOPATH=/winesnob/go
RUN go get github.com/gorilla/mux \
    github.com/gorilla/handlers \
    github.com/jinzhu/gorm \
    github.com/jinzhu/gorm/dialects/sqlite

# COPY dymo-labelgen ./dymo-labelgen
# RUN pip install -r dymo-labelgen/requirements.txt

# COPY package.json ./
# RUN npm install

COPY wine-list.json ./

WORKDIR go/src/
COPY go/src/backend ./backend

# copy frontend files and compile, resulting in a statically-servable "dist" directory
# COPY protractor.conf.js tslint.json karma.conf.js angular-cli.json ./
# COPY src ./src
# TODO: fix ng build
# RUN ng build --env=prod || true

VOLUME "/data"
EXPOSE 8080
CMD ["go", "run", "backend/main.go", "--dbpath", "/data/parts.sqlite3db"]
