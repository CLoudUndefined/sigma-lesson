from aiogram.filters import Command
from aiogram.types import Message

from ..database.repository import repository

from .admin_router import admin_router


@admin_router.message(Command("free"))
async def free(message: Message) -> None:
    parts = message.text.split(maxsplit=1)

    if len(parts) != 2:
        await message.answer(
            "Использование:\n"
            "<code>/free @username</code>\n"
            "или\n"
            "<code>/free Иванов Иван Иванович</code>"
        )
        return

    query = parts[1].strip()

    if query.startswith("@"):
        user = await repository.get_user_by_username(
            query[1:]
        )
    else:
        user = await repository.get_user_by_full_name(
            query
        )

    if user is None:
        await message.answer(
            "Пользователь не найден."
        )
        return

    access = await repository.get_user_access(
        user.telegram_id
    )

    if access is None:
        await message.answer(
            "За пользователем не закреплен сервер."
        )
        return

    await repository.release_access(access.id)

    username = (
        f"@{user.username}"
        if user.username
        else "—"
    )

    await message.answer(
        "Доступ освобожден.\n\n"
        f"Пользователь: {username}\n"
        f"ФИО: {user.full_name}\n"
        f"Логин: <code>{access.login}</code>\n"
        f"Порт: <code>{access.port}</code>"
    )
