from aiogram.filters import Command
from aiogram.types import Message

from ..database.repository import repository
from .admin_router import admin_router


@admin_router.message(Command("find"))
async def find(message: Message) -> None:
    parts = message.text.split(maxsplit=1)

    if len(parts) != 2:
        await message.answer(
            "Использование:\n"
            "<code>/find @username</code>"
        )
        return

    username = parts[1].strip()

    if not username.startswith("@"):
        await message.answer(
            "Неправильный формат.\n"
            "Укажите пользователя в формате <code>@username</code>."
        )
        return

    user = await repository.get_user_details_by_username(
        username[1:]
    )

    if user is None:
        await message.answer(
            "Пользователь не найден."
        )
        return

    lines = [
        "<b>Информация о пользователе</b>",
        "",
        f"Username: @{user.username}",
        f"ФИО: {user.full_name}",
        f"Telegram ID: <code>{user.telegram_id}</code>",
        f"Дата регистрации: {user.registered_at}",
        "",
        "<b>Доступ к серверу</b>",
    ]

    if user.login is None:
        lines.append("Не закреплен ни за одним сервером.")
    else:
        lines.extend([
            f"Login: <code>{user.login}</code>",
            f"Password: <code>{user.password}</code>",
            f"Port: <code>{user.port}</code>",
            f"Domain: <code>{user.domain}</code>",
        ])

    await message.answer("\n".join(lines))
