-- Script Inisialisasi Basis Data Terisolasi Microservices BangunGrid
CREATE DATABASE db_auth;
CREATE DATABASE db_catalog;
CREATE DATABASE db_inventory;
CREATE DATABASE db_order;
CREATE DATABASE db_fleet;
CREATE DATABASE db_billing;
CREATE DATABASE db_notification;

-- Aktifkan PostGIS pada basis data fleet
\c db_fleet
CREATE EXTENSION IF NOT EXISTS postgis;
