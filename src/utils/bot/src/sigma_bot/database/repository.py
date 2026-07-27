from .db import database
from .models import Access, User, AccessInfo, UserInfo


class Repository:
    async def create_user(
        self,
        telegram_id: int,
        username: str | None,
        full_name: str,
    ) -> None:
        await database.connection.execute(
            """
            INSERT INTO users (
                telegram_id,
                username,
                full_name
            )
            VALUES (?, ?, ?)
            """,
            (
                telegram_id,
                username,
                full_name,
            ),
        )

        await database.connection.commit()

    async def delete_user(
        self,
        telegram_id: int,
    ) -> None:
        await database.connection.execute(
            """
            DELETE FROM users
            WHERE telegram_id = ?
            """,
            (telegram_id,),
        )

        await database.connection.commit()

    async def user_exists(
        self,
        telegram_id: int,
    ) -> bool:
        cursor = await database.connection.execute(
            """
            SELECT 1
            FROM users
            WHERE telegram_id = ?
            LIMIT 1
            """,
            (telegram_id,),
        )

        return await cursor.fetchone() is not None

    async def get_user(
        self,
        telegram_id: int,
    ) -> User | None:
        cursor = await database.connection.execute(
            """
            SELECT
                telegram_id,
                username,
                full_name,
                registered_at
            FROM users
            WHERE telegram_id = ?
            LIMIT 1
            """,
            (telegram_id,),
        )

        row = await cursor.fetchone()

        if row is None:
            return None

        return User(*row)

    async def get_user_access(
        self,
        telegram_id: int,
    ) -> Access | None:
        cursor = await database.connection.execute(
            """
            SELECT
                id,
                login,
                password,
                port,
                domain,
                telegram_id
            FROM accesses
            WHERE telegram_id = ?
            LIMIT 1
            """,
            (telegram_id,),
        )

        row = await cursor.fetchone()

        if row is None:
            return None

        return Access(*row)

    async def get_free_access(
        self,
    ) -> Access | None:
        cursor = await database.connection.execute(
            """
            SELECT
                id,
                login,
                password,
                port,
                domain,
                telegram_id
            FROM accesses
            WHERE telegram_id IS NULL
            ORDER BY id
            LIMIT 1
            """
        )

        row = await cursor.fetchone()

        if row is None:
            return None

        return Access(*row)

    async def assign_access(
        self,
        access_id: int,
        telegram_id: int,
    ) -> None:
        await database.connection.execute(
            """
            UPDATE accesses
            SET telegram_id = ?
            WHERE id = ?
            """,
            (
                telegram_id,
                access_id,
            ),
        )

        await database.connection.commit()

    async def release_access(
        self,
        access_id: int,
    ) -> None:
        await database.connection.execute(
            """
            UPDATE accesses
            SET telegram_id = NULL
            WHERE id = ?
            """,
            (access_id,),
        )

        await database.connection.commit()

    async def get_all_users(
        self,
    ) -> list[UserInfo]:
        cursor = await database.connection.execute(
            """
            SELECT
                users.username,
                users.full_name,
                accesses.login
            FROM users
            LEFT JOIN accesses
                ON accesses.telegram_id = users.telegram_id
            ORDER BY users.full_name;
            """
        )

        rows = await cursor.fetchall()

        return [
            UserInfo(*row)
            for row in rows
        ]

    async def get_all_accesses(self) -> list[AccessInfo]:
        cursor = await database.connection.execute(
            """
            SELECT
                accesses.login,
                accesses.password,
                accesses.port,
                users.username
            FROM accesses
            LEFT JOIN users
                ON users.telegram_id = accesses.telegram_id
            ORDER BY accesses.id
            """
        )

        rows = await cursor.fetchall()

        return [
            AccessInfo(*row)
            for row in rows
        ]

    async def get_user_by_username(
        self,
        username: str,
    ) -> User | None:
        cursor = await database.connection.execute(
            """
            SELECT
                telegram_id,
                username,
                full_name,
                registered_at
            FROM users
            WHERE username = ?
            LIMIT 1
            """,
            (username,),
        )

        row = await cursor.fetchone()

        if row is None:
            return None

        return User(*row)

    async def get_user_by_full_name(
        self,
        full_name: str,
    ) -> User | None:
        cursor = await database.connection.execute(
            """
            SELECT
                telegram_id,
                username,
                full_name,
                registered_at
            FROM users
            WHERE full_name = ?
            LIMIT 1
            """,
            (full_name,),
        )

        row = await cursor.fetchone()

        if row is None:
            return None

        return User(*row)
    
repository = Repository()
