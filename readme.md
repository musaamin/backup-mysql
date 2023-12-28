# MySQL/MariaDB Backup Tool

Program sederhana untuk melakukan backup MySQL/MariaDB database dan sinkronisasi direktori backup ke cloud storage. Program ini memanfaatkan `mysqldump` untuk export database dan `rclone` untuk sinkronisasi ke cloud.

## Cara Install
Download `backup-mysql`

```
sudo wget https://raw.githubusercontent.com/musaamin/backup-mysql/main/backup-mysql -O /usr/local/bin/backup-mysql
sudo chmod +x /usr/local/bin/backup-mysql
```

## Cara Menggunakan
Membuat file konfigurasi yang berisi informasi database dan backup.

```
backup-mysql init mydb.txt
```

Atur file konfigurasi `mydb.txt`.

```
DBHOST=localhost
DBPORT=3306
DBNAME=mydatabase
DBUSER=myuser
DBPASS=mypassword
BACKUPDIR=/path/to/backup
RCLONEREMOTES=remote1, remote2
RCLONEDIR=cloud-dir/backup
```

Melakukan backup database.

```
backup-mysql export mydb.txt
```

Melakukan sinkronisasi direktori backup ke cloud storage.

```
backup-mysql rclone mydb.txt
```