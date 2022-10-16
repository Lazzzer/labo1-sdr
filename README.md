# Laboratoire 1 de SDR

## Auteurs

Lazar Pavicevic et Jonathan Friedli

## Contexte

Ce projet est réalisé dans le cadre du cours de Systèmes Distribués et Répartis (SDR) de la HEIG-VD. Il a pour but de mettre en place un système de gestion de manifestation via une application client-serveur TCP-IP. Le créateur de la manifestation pourra créer différents jobs et les bénévoles pourront s'y inscrire.

## Utilisation du programme

Le programme contient plusieurs flags qui permettent de choisir le lancement d'un serveur ou d'un client. Ils spécifient aussi si le serveur doit être lancé en mode `debug` et/ou en mode `silent`.

Le mode `debug` ralentit artificiellement de 4 secondes le serveur lorsqu'il rentre dans des sections critiques et affiche des messages d'entrées/sorties de ces dernières.

Le mod `silent` désactive les logs du serveur. Cependant, les logs du mode debug sont toujours affichés. Ceci est surtout pratique pour observer le comportement du serveur lor des tests d'intégration.

### Une commande pour connaître les usages est disponible :

```bash
# A la racine du projet
go run . -h
# Ou
go run .\main.go --help

# Ou si le projet a été compilé et que l'exécutable se trouve dans le dossier courant
.\labo1-sdr.exe --help
```

Résultat:

```bash
Usage of labo1-sdr.exe:
  -debug
        Boolean: Run server in debug mode. Default is false
  -server
        Boolean: Run program in server mode. Default is client mode
  -silent
        Boolean: Run server in silent mode. Default is false
```

### Pour lancer un client:

```bash
# A la racine du projet
go run .\main.go

# Ou en mode race
go run -race .\main.go
```

### Pour lancer un serveur:

```bash
# A la racine du projet
go run .\main.go --server

# En mode race & debug
go run -race .\main.go --server --debug

# En mode silent
go run .\main.go --server --silent
```

### Pour lancer les tests automatisés:

```bash
# A la racine du projet
go test -race .\client -v

# Si besoins, en vidant le cache
go clean -testcache && go test -race .\client -v
```

## Liste des commandes utilisateur

```bash
# Afficher de l'aide
help

# Créer une manifestation (Demande le nom d'utilisateur et le mot de passe de l'utilisateur)
create <eventName> <jobName1> <nbVolunteer1> [<jobName2> <nbVolunteer2>...] [[<username> <password>]]

# Clore une manifestation (Demande le nom d'utilisateur et le mot de passe de l'utilisateur)
close <idEvent> [[<username> <password>]]

# S'inscrire à une manifestation (Demande le nom d'utilisateur et le mot de passe de l'utilisateur)
register <idEvent> <idJob> [[<username> <password>]]

# Afficher toutes les manifestations ou une manifestation spécifique
show [<idEvent>]

# Afficher les bénévoles d'une certaine manifestation
jobs <idEvent>

# Quitter le programme
quit
```
