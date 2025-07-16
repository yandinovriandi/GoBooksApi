# 📚 Go Books API

Ini adalah RESTful API sederhana yang dibangun dengan Go dan Gin framework untuk mengelola koleksi buku. API ini menyediakan operasi CRUD (Create, Read, Update, Delete) untuk data buku, dengan validasi input dan penanganan error yang ramah.

---

## ✨ Fitur

* **Tambah Buku**: Membuat entri buku baru dengan judul, penulis, dan tahun publikasi.
* **Lihat Semua Buku**: Mengambil daftar semua buku yang tersedia.
* **Lihat Buku Berdasarkan ID**: Mengambil detail satu buku berdasarkan ID uniknya.
* **Perbarui Buku**: Memperbarui informasi buku yang sudah ada.
* **Hapus Buku**: Menghapus buku dari koleksi.
* **Validasi Input**: Memastikan data yang diterima sesuai format yang diharapkan.
* **Penanganan Error**: Respons error yang informatif untuk berbagai skenario (input tidak valid, buku tidak ditemukan, error server).

---

## 🚀 Teknologi yang Digunakan

* **Go**: Bahasa pemrograman utama.
* **Gin Web Framework**: Framework web cepat untuk membangun API.
* **`database/sql`**: Paket standar Go untuk interaksi database.
* **`go-playground/validator`**: Pustaka untuk validasi *struct*.
* **Database**: SQLite (contoh, bisa diganti dengan PostgreSQL, MySQL, dll.).

---

## 📋 Prasyarat

Sebelum menjalankan proyek ini, pastikan Anda memiliki:

* **Go** (versi 1.16 atau lebih baru) terinstal di sistem Anda.
* **Database** (misalnya SQLite, yang akan dibuat secara otomatis jika Anda menggunakan `database/sqlite.go` seperti contoh).

---

## 📦 Struktur Proyek

go-book-api/
├── main.go               # Logika utama aplikasi dan definisi endpoint
├── database/
│   └── database.go       # Inisialisasi koneksi database dan skema
└── models/
└── book.go           # Definisi struktur data Book


---

## ⚙️ Instalasi & Setup

1.  **Kloning Repositori:**

    ```bash
    git clone [https://github.com/yandinovriandi/GoBooksApi.git](https://github.com/yandinovriandi/GoBooksApi.git)
    cd go-book-api
    ```

2.  **Unduh Dependensi:**

    ```bash
    go mod tidy
    ```

3.  **Konfigurasi Database:**
    Asumsi Anda memiliki file `database/database.go` yang menginisialisasi database dan membuat tabel `books`. Contoh sederhana untuk SQLite:

    ```go
    // database/database.go
    package database

    import (
        "database/sql"
        _ "[github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)" // Driver SQLite
        "log"
    )

    var DB *sql.DB

    func InitDB() {
        var err error
        DB, err = sql.Open("sqlite3", "./books.db") // Membuat file books.db
        if err != nil {
            log.Fatal(err)
        }

        // Pastikan koneksi berfungsi
        err = DB.Ping()
        if err != nil {
            log.Fatal(err)
        }

        createTableSQL := `
        CREATE TABLE IF NOT EXISTS books (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            author TEXT NOT NULL,
            publication_year INTEGER NOT NULL
        );`

        _, err = DB.Exec(createTableSQL)
        if err != nil {
            log.Fatal(err)
        }
        log.Println("Database dan tabel 'books' siap.")
    }
    ```

    Dan file `models/book.go`:

    ```go
    // models/book.go
    package models

    type Book struct {
        ID              int    `json:"id"`
        Title           string `json:"title" binding:"required"`
        Author          string `json:"author" binding:"required"`
        PublicationYear int    `json:"publication_year" binding:"required"`
    }
    ```

4.  **Jalankan Aplikasi:**

    ```bash
    go run main.go
    ```

    Aplikasi akan berjalan di `http://localhost:8080`.

---

## 📚 API Endpoints

Berikut adalah daftar *endpoint* yang tersedia:

### 1. `POST /books` - Menambahkan Buku Baru

* **Deskripsi**: Membuat entri buku baru di database.
* **Metode**: `POST`
* **URL**: `http://localhost:8080/books`
* **Headers**:
   * `Content-Type: application/json`
* **Request Body (JSON)**:
    ```json
    {
        "title": "Judul Buku Baru",
        "author": "Nama Penulis",
        "publication_year": 2023
    }
    ```
* **Contoh Respons Sukses (Status: 201 Created)**:
    ```json
    {
        "id": 1,
        "title": "Judul Buku Baru",
        "author": "Nama Penulis",
        "publication_year": 2023
    }
    ```
* **Contoh Respons Error (Validasi, Status: 400 Bad Request)**:
    ```json
    {
        "error": "field_validation",
        "message": "Beberapa field wajib di isi",
        "fields": {
            "Title": "Title wajib di isi",
            "PublicationYear": "PublicationYear wajib di isi"
        }
    }
    ```

### 2. `GET /books` - Mendapatkan Semua Buku

* **Deskripsi**: Mengambil daftar semua buku yang ada di database.
* **Metode**: `GET`
* **URL**: `http://localhost:8080/books`
* **Contoh Respons Sukses (Status: 200 OK)**:
    ```json
    {
        "message": "Data buku berhasil diambil",
        "data": [
            {
                "id": 1,
                "title": "Judul Buku Pertama",
                "author": "Penulis A",
                "publication_year": 2020
            },
            {
                "id": 2,
                "title": "Judul Buku Kedua",
                "author": "Penulis B",
                "publication_year": 2021
            }
        ],
        "total": 2
    }
    ```
* **Contoh Respons Tanpa Data (Status: 200 OK)**:
    ```json
    {
        "message": "Tidak ada buku ditemukan.",
        "data": [],
        "total": 0
    }
    ```

### 3. `GET /books/:id` - Mendapatkan Buku Berdasarkan ID

* **Deskripsi**: Mengambil detail satu buku berdasarkan ID uniknya.
* **Metode**: `GET`
* **URL**: `http://localhost:8080/books/{id}` (contoh: `http://localhost:8080/books/1`)
* **Contoh Respons Sukses (Status: 200 OK)**:
    ```json
    {
        "id": 1,
        "title": "Judul Buku Pertama",
        "author": "Penulis A",
        "publication_year": 2020
    }
    ```
* **Contoh Respons Error (Buku Tidak Ditemukan, Status: 404 Not Found)**:
    ```json
    {
        "message": "Buku tidak ditemukan"
    }
    ```
* **Contoh Respons Error (ID Tidak Valid, Status: 400 Bad Request)**:
    ```json
    {
        "error": "ID buku tidak valid"
    }
    ```

### 4. `PUT /books/:id` - Memperbarui Buku Berdasarkan ID

* **Deskripsi**: Memperbarui informasi buku yang sudah ada berdasarkan ID.
* **Metode**: `PUT`
* **URL**: `http://localhost:8080/books/{id}` (contoh: `http://localhost:8080/books/1`)
* **Headers**:
   * `Content-Type: application/json`
* **Request Body (JSON)**:
    ```json
    {
        "title": "Judul Buku Diperbarui",
        "author": "Penulis Diperbarui",
        "publication_year": 2024
    }
    ```
* **Contoh Respons Sukses (Status: 200 OK)**:
    ```json
    {
        "id": 1,
        "title": "Judul Buku Diperbarui",
        "author": "Penulis Diperbarui",
        "publication_year": 2024
    }
    ```
* **Contoh Respons Error (Buku Tidak Ditemukan, Status: 404 Not Found)**:
    ```json
    {
        "message": "Buku tidak ditemukan"
    }
    ```
* **Contoh Respons Error (Validasi, Status: 400 Bad Request)**:
    ```json
    {
        "error": "field_validation",
        "message": "Beberapa field wajib di isi atau tidak valid",
        "fields": {
            "Title": "Title wajib di isi"
        }
    }
    ```

### 5. `DELETE /books/:id` - Menghapus Buku Berdasarkan ID

* **Deskripsi**: Menghapus buku dari database berdasarkan ID.
* **Metode**: `DELETE`
* **URL**: `http://localhost:8080/books/{id}` (contoh: `http://localhost:8080/books/7`)
* **Contoh Respons Sukses (Status: 200 OK)**:
    ```json
    {
        "message": "Buku berhasil dihapus"
    }
    ```
* **Contoh Respons Error (Buku Tidak Ditemukan, Status: 404 Not Found)**:
    ```json
    {
        "message": "Buku tidak ditemukan"
    }
    ```
* **Contoh Respons Error (ID Tidak Valid, Status: 400 Bad Request)**:
    ```json
    {
        "error": "ID buku tidak valid"
    }
    ```

---

## 🤝 Kontribusi

Jika Anda ingin berkontribusi pada proyek ini, silakan *fork* repositori, buat cabang baru, dan kirim *pull request* Anda.

---

## 📄 Lisensi

Proyek ini dilisensikan di bawah Lisensi MIT. Lihat file [LICENSE](https://github.com/yandinovriandi/GoBooksApi/blob/main/LICENSE) untuk detail lebih lanjut.