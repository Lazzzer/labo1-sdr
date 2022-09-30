# Laboratoire 1 de SDR

## Auteurs
Lazar Pavicevic et Jonathan Friedli

## Contexte
Ce projet est réalisé dans le cadre du cours de Systèmes Distribués et Répartis (SDR) de la HEIG-VD. Il a pour but de mettre en place un système de gestion de manifestation. Le créateur de la manifestation pourra créer différents job et les bénévoles pourront s'y inscrire.

## Liste des commandes utilisateur
```bash
# Afficher de l'aide
help

# Quitter le programme
quit

# Créer une manifestation
create <name> <username> <password> <job1> <nbVolunteer1> [<job2> <nbVolunteer2> ...]

# Clore une manifestation
close <idManifestation> <username> <password>

# S'inscrire à une manifestation
register <idManifestation> <idJob> <username> <password> 

# Afficher toutes les manifestations
showAll

# Afficher les job d'une certaine manifestation
showJobs <idManifestation>

# Afficher les bénévoles d'une certaine manifestation
jobRepartition <idManifestation>
```