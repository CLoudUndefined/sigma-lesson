import csv

from pathlib import Path

import aiosqlite

ROOT_DIR = Path(__file__).resolve().parents[3]
DATA_DIR = ROOT_DIR / "data"
DB_PATH = DATA_DIR / "database.db"


def load_accesses() -> list[tuple[str, str, int, str]]:
    csv_path = DATA_DIR / "students.csv"

    if not csv_path.exists():
        raise FileNotFoundError(
            f"File not found: {csv_path}"
        )

    with open(
        csv_path,
        newline="",
        encoding="utf-8",
    ) as file:
        reader = csv.DictReader(file)

        accesses = [
            (
                row["login"],
                row["password"],
                int(row["port"]),
                row["domain"],
            )
            for row in reader
        ]

    if not accesses:
        raise RuntimeError(
            "students.csv is empty."
        )

    return accesses


class Database:
    def __init__(self):
        self._connection: aiosqlite.Connection | None = None

    async def connect(self) -> None:
        DATA_DIR.mkdir(exist_ok=True)

        self._connection = await aiosqlite.connect(DB_PATH)
        await self._connection.execute("PRAGMA foreign_keys = ON;")

    async def close(self) -> None:
        if self._connection is not None:
            await self._connection.close()
            self._connection = None

    @property
    def connection(self) -> aiosqlite.Connection:
        if self._connection is None:
            raise RuntimeError(
                "Database is not connected."
            )

        return self._connection

    async def initialize(self) -> None:
        await self.connection.execute("""
            CREATE TABLE IF NOT EXISTS users (
                telegram_id INTEGER PRIMARY KEY,
                username TEXT,
                full_name TEXT NOT NULL,
                registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        """)

        await self.connection.execute("""
            CREATE TABLE IF NOT EXISTS accesses (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                login TEXT NOT NULL UNIQUE,
                password TEXT NOT NULL,
                port INTEGER NOT NULL UNIQUE,
                domain TEXT NOT NULL UNIQUE,
                telegram_id INTEGER UNIQUE,

                FOREIGN KEY (telegram_id)
                    REFERENCES users (telegram_id)
                    ON DELETE SET NULL
            )
        """)

        await self.connection.executemany(
            """
            INSERT OR IGNORE INTO accesses (
                login,
                password,
                port,
                domain
            )
            VALUES (?, ?, ?, ?)
            """,
            load_accesses(),
        )

        await self.connection.commit()


database = Database()
