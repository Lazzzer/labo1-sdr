# Laboratoire 1 de SDR

## Auteurs

Lazar Pavicevic et Jonathan Friedli

## Contexte

Ce projet est réalisé dans le cadre du cours de Systèmes Distribués et Répartis (SDR) de la HEIG-VD. Il a pour but de mettre en place un système de gestion de manifestation via une application client-serveur TCP-IP. Le créateur de la manifestation pourra créer différents jobs et les bénévoles pourront s'y inscrire.

## Liste des commandes utilisateur

```bash
# Afficher de l'aide
help

# Créer une manifestation (Demande le mot de passe de l'utilisateur)
create <name> <username> <job1> <nbVolunteer1> [<job2> <nbVolunteer2> ...]

# Clore une manifestation (Demande le mot de passe de l'utilisateur)
close <idEvent> <username>

# S'inscrire à une manifestation (Demande le mot de passe de l'utilisateur)
register <idEvent> <idJob> <username> [[<password>]]

# Afficher toutes les manifestations
showAll

# Afficher les job d'une certaine manifestation
showJobs <idEvent>

# Afficher les bénévoles d'une certaine manifestation
jobRepartition <idEvent>

# Quitter le programme
quit
```
