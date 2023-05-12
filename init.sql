CREATE DATABASE IF NOT EXISTS `server`;

USE `server`;

CREATE TABLE IF NOT EXISTS `consumption` (
                                             `uuid` BINARY(16),
                                             `time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                             `userid` BINARY(16) NOT NULL,
                                             `postalcode` VARCHAR(255) NOT NULL,
                                             `city` VARCHAR(255) NOT NULL,
                                             `address` VARCHAR(255) NOT NULL,
                                             `wamount` INT NOT NULL,
                                             PRIMARY KEY (uuid, time)
);

CREATE TABLE IF NOT EXISTS `feeding` (
                                         `uuid` BINARY(16),
                                         `time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                         `userid` BINARY(16) NOT NULL,
                                         `postalcode` VARCHAR(255) NOT NULL,
                                         `city` VARCHAR(255) NOT NULL,
                                         `address` VARCHAR(255) NOT NULL,
                                         `powerstorage` BINARY(1) NOT NULL,
                                         `pscapacity` DEC NOT NULL,
                                         `wamount` INT NOT NULL,
                                         PRIMARY KEY (uuid, time)
);

CREATE TABLE IF NOT EXISTS `changes` (
                                         `uuid` BINARY(16),
                                         `time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                         `userid` BINARY(16) NOT NULL,
                                         `absVal` INT NOT NULL,
                                         `wChange` INT NOT NULL,
                                         PRIMARY KEY (uuid, time)
);

# ALTER TABLE `changes`
#     ADD FOREIGN KEY (`userid`)
#         REFERENCES `consumption` (`userid`),
#     ADD FOREIGN KEY (`userid`)
#         REFERENCES `feeding` (`userid`);

# ALTER TABLE `transactions` ADD FOREIGN KEY (`seller`) REFERENCES `users` (`uuid`);
#
# ALTER TABLE `transactions` ADD FOREIGN KEY (`buyer`) REFERENCES `users` (`uuid`);
#
# ALTER TABLE `energy` ADD FOREIGN KEY (`user`) REFERENCES `users` (`uuid`);
#
# ALTER TABLE `power` ADD FOREIGN KEY (`user`) REFERENCES `users` (`uuid`);
#
# -- example users for matching --
# INSERT INTO users (uuid, name, seller, public_key, plz, email, password, iban, chainid, active, minprice, maxprice, matched)
# -- sellers -
# VALUES ((UUID_TO_BIN('249d442a-d5fc-4c65-bb43-7f21d7eecfd4', true)), 'Max Mustermann', 1, 'dummypublickey', 81541, 'max@mustermann.de', 'pwhash', 'DE20500105172818255489', 'maxchain', 1, 14, 0, 0),
#        ((UUID_TO_BIN('44146e78-0d7a-4c2f-b2e4-f3fcccd29175', true)), 'Robert Frank', 1, 'dummypublickey', 81541, 'robert.frank@tum.de', 'pwhash', 'DE12500105171854317939', 'robertchain', 1, 10, 0, 0),
#        ((UUID_TO_BIN('7d61316c-1456-452c-ac41-a373756b38a3', true)), 'Rudolf Bayer', 1, 'dummypublickey', 81541, 'bayerr@in.tum.de', 'pwhash', 'DE27500105178692893911', 'rudolfchain', 1, 17, 0, 0),
#        ((UUID_TO_BIN('153f948a-fa4a-446f-aa13-fa70be3f15fb', true)), 'Peter Breitschaft', 1, 'dummypublickey', 81541, 'peter.breitschaft@googlemail.com', 'pwhash', 'DE10500105179441579696', 'peterchain', 1, 18, 0, 0),
#        ((UUID_TO_BIN('03aa9c1b-49db-4f19-9fda-c46928afe320', true)), 'Jan Seidemann', 1, 'dummypublickey', 81541, 'jan.seidemann@fenecon.de', 'pwhash', 'DE98500105173121185279', 'janchain', 1, 12, 0, 0),
#        ((UUID_TO_BIN('8a970131-b6e1-4767-a446-225fdf89d3e1', true)), 'Daniel Birkeneder', 1, 'dummypublickey', 81541, 'daniel.birkeneder@fenecon.de', 'pwhash', 'DE94500105175466162582', 'danielchain', 1, 13, 0, 0),
# -- buyers --
#        ((UUID_TO_BIN('1865ce5f-f865-43ba-bd9c-5c18f7f4f17e', true)), 'Angela Adams', 0, 'dummypublickey', 81541, 'angela@adams.com', 'pwhash', 'DE46500105179853842725', 'angelachain', 1, 0, 14, 0),
#        ((UUID_TO_BIN('f0676026-f7e0-40d8-9102-79beb81716c4', true)), 'Bertha Butler', 0, 'dummypublickey', 81541, 'bertha.butler@yahoo.de', 'pwhash', 'DE25500105174978456651', 'berthachain', 1, 0, 12, 0),
#        ((UUID_TO_BIN('ac3c7dea-9093-4b0c-90b1-b3350322d673', true)), 'Chloe Carter', 0, 'dummypublickey', 81541, 'c@carter.xyz', 'pwhash', 'DE14500105177757143778', 'chloechain', 1, 0, 20, 0),
#        ((UUID_TO_BIN('3de6fe02-ca22-48f9-8735-b093797e4e53', true)), 'Daisy Duck', 0, 'dummypublickey', 81541, 'daisy@duck.com', 'pwhash', 'DE12500105171896827898', 'daisychain', 1, 0, 15, 0),
#        ((UUID_TO_BIN('470236a9-0045-4798-8f0b-1a6a73cfdce4', true)), 'Emma East', 0, 'dummypublickey', 81541, 'emma@eastwards.de', 'pwhash', 'DE36500105179468737773', 'emmachain', 1, 0, 18, 0),
#        ((UUID_TO_BIN('d23b4d4f-78ea-4d9b-af7b-18e25c020fcc', true)), 'Francesca Fallon', 0, 'dummypublickey', 81541, 'fran.fallon@icloud.com', 'pwhash', 'DE40500105175553553958', 'franchain', 1, 0, 15, 0),
#        ((UUID_TO_BIN('0d0c672a-bc57-40cb-8cc7-d2596bfddab0', true)), 'Gabriele Gartner', 0, 'dummypublickey', 81541, 'gaby@gartner.com', 'pwhash', 'DE03500105177857943466', 'gabrielechain', 1, 0, 15, 0);
#
# -- example users for querying --
# INSERT INTO users (uuid, name, seller, public_key, plz, email, password, iban, chainid, active, minprice, maxprice, matched, url)
# VALUES ((UUID_TO_BIN('4bea705a-3d02-4681-a1cf-c3ef1164adcb', true)), 'Victor Venus', 1, 'dummypublickey', 80333, 'victor@venus.io', 'Basic YWRtaW46YWRtaW4=','DE10500105179771128466', 'victorchain', 1, 22, 0, 0, 'https://8084-openems-openems-ithyl4oh5hq.ws-eu71.gitpod.io/rest/channel/'),
#        ((UUID_TO_BIN('a4a1c320-6157-438e-8311-e33e24641e2d', true)), 'Whiyney Walter', 0, 'dummypublickey', 80333, 'whitneyw@web.de', 'pwhash', 'DE66500105171278249681', 'whitneychain', 1, 0, 26, 0, 'http://192.168.10.5:8081');
