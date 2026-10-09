#include <chrono>
#include <cstdio>
#include <cstdlib>
#include <filesystem>
#include <iostream>
#include <string>

#include <rocksdb/db.h>
#include <rocksdb/options.h>

namespace fs = std::filesystem;

constexpr int NUM_KEYS = 1000;
constexpr int NUM_OPS = 10000;

std::string Key(int i) {
    char buffer[32];
    std::snprintf(buffer, sizeof(buffer), "key-%04d", i);
    return std::string(buffer);
}

std::string Value(int i) {
    char buffer[32];
    std::snprintf(buffer, sizeof(buffer), "value-%04d", i);
    return std::string(buffer);
}

rocksdb::DB* CreateDB(const std::string& path) {
    rocksdb::Options options;
    options.create_if_missing = true;

    rocksdb::DB* db = nullptr;

    rocksdb::Status status =
        rocksdb::DB::Open(options, path, &db);

    if (!status.ok()) {
        std::cerr << "Failed to open RocksDB: "
                  << status.ToString() << std::endl;
        return nullptr;
    }

    return db;
}

void PrepareData(rocksdb::DB* db) {
    rocksdb::WriteOptions write_options;

    for (int i = 0; i < NUM_KEYS; ++i) {
        std::string key = Key(i);
        std::string value = Value(i);

        rocksdb::Status status =
            db->Put(write_options, key, value);

        if (!status.ok()) {
            std::cerr << "Setup PUT failed: "
                      << status.ToString() << std::endl;
            std::exit(1);
        }
    }
}

double BenchmarkPut() {
    std::string path = "/tmp/resilientkv-rocksdb-put";

    fs::remove_all(path);

    rocksdb::DB* db = CreateDB(path);

    if (!db) {
        std::exit(1);
    }

    rocksdb::WriteOptions write_options;

    auto start = std::chrono::steady_clock::now();

    for (int i = 0; i < NUM_OPS; ++i) {
        std::string key = Key(i % NUM_KEYS);

        rocksdb::Status status =
            db->Put(write_options, key, "updated");

        if (!status.ok()) {
            std::cerr << "PUT failed: "
                      << status.ToString() << std::endl;
            delete db;
            std::exit(1);
        }
    }

    auto end = std::chrono::steady_clock::now();

    delete db;
    fs::remove_all(path);

    double seconds =
        std::chrono::duration<double>(end - start).count();

    return (seconds * 1e9) / NUM_OPS;
}

double BenchmarkGet() {
    std::string path = "/tmp/resilientkv-rocksdb-get";

    fs::remove_all(path);

    rocksdb::DB* db = CreateDB(path);

    if (!db) {
        std::exit(1);
    }

    PrepareData(db);

    rocksdb::ReadOptions read_options;
    std::string value;

    auto start = std::chrono::steady_clock::now();

    for (int i = 0; i < NUM_OPS; ++i) {
        std::string key = Key(i % NUM_KEYS);

        rocksdb::Status status =
            db->Get(read_options, key, &value);

        if (!status.ok()) {
            std::cerr << "GET failed: "
                      << status.ToString() << std::endl;
            delete db;
            std::exit(1);
        }
    }

    auto end = std::chrono::steady_clock::now();

    delete db;
    fs::remove_all(path);

    double seconds =
        std::chrono::duration<double>(end - start).count();

    return (seconds * 1e9) / NUM_OPS;
}

double BenchmarkDelete() {
    std::string path = "/tmp/resilientkv-rocksdb-delete";

    fs::remove_all(path);

    rocksdb::DB* db = CreateDB(path);

    if (!db) {
        std::exit(1);
    }

    PrepareData(db);

    rocksdb::WriteOptions write_options;

    auto start = std::chrono::steady_clock::now();

    for (int i = 0; i < NUM_OPS; ++i) {
        std::string key = Key(i % NUM_KEYS);

        rocksdb::Status status =
            db->Delete(write_options, key);

        if (!status.ok()) {
            std::cerr << "DELETE failed: "
                      << status.ToString() << std::endl;
            delete db;
            std::exit(1);
        }
    }

    auto end = std::chrono::steady_clock::now();

    delete db;
    fs::remove_all(path);

    double seconds =
        std::chrono::duration<double>(end - start).count();

    return (seconds * 1e9) / NUM_OPS;
}

double BenchmarkMixed() {
    std::string path = "/tmp/resilientkv-rocksdb-mixed";

    fs::remove_all(path);

    rocksdb::DB* db = CreateDB(path);

    if (!db) {
        std::exit(1);
    }

    PrepareData(db);

    rocksdb::WriteOptions write_options;
    rocksdb::ReadOptions read_options;

    std::string value;

    auto start = std::chrono::steady_clock::now();

    for (int i = 0; i < NUM_OPS; ++i) {
        std::string key = Key(i % NUM_KEYS);

        if (i % 3 == 0) {
            db->Put(
                write_options,
                key,
                "updated"
            );
        }
        else if (i % 3 == 1) {
            db->Get(
                read_options,
                key,
                &value
            );
        }
        else {
            db->Put(
                write_options,
                key,
                "updated-again"
            );
        }
    }

    auto end = std::chrono::steady_clock::now();

    delete db;
    fs::remove_all(path);

    double seconds =
        std::chrono::duration<double>(end - start).count();

    return (seconds * 1e9) / NUM_OPS;
}

int main() {
    std::cout << "RocksDB Benchmark" << std::endl;
    std::cout << "=================" << std::endl;
    std::cout << "RocksDB version: 8.9.1" << std::endl;
    std::cout << "Operations per benchmark: "
              << NUM_OPS << std::endl;
    std::cout << "Dataset size: "
              << NUM_KEYS << " keys" << std::endl;
    std::cout << std::endl;

    std::cout << "PUT:    "
              << BenchmarkPut()
              << " ns/op" << std::endl;

    std::cout << "GET:    "
              << BenchmarkGet()
              << " ns/op" << std::endl;

    std::cout << "DELETE: "
              << BenchmarkDelete()
              << " ns/op" << std::endl;

    std::cout << "MIXED:  "
              << BenchmarkMixed()
              << " ns/op" << std::endl;

    return 0;
}
