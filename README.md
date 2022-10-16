# Laboratoire 1 de SDR

## Auteurs

Lazar Pavicevic et Jonathan Friedli

## Contexte

Ce projet est réalisé dans le cadre du cours de Systèmes Distribués et Répartis (SDR) de la HEIG-VD. Il a pour but de mettre en place un système de gestion de manifestation via une application client-serveur TCP-IP. Le créateur de la manifestation pourra créer différents jobs et les bénévoles pourront s'y inscrire.

## Liste des commandes utilisateur

```bash
# Afficher de l'aide
help

# Créer une manifestation (Demande le nom d'utilisateur et le mot de passe de l'utilisateur)
create <name> <jobName1> <nbVolunteer1> [<jobName2> <nbVolunteer2> ...] [[<username> <password>]]

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
