from aiogram.filters import Command
from aiogram.types import Message

from ..database.repository import repository
from .user_router import user_router


@user_router.message(Command("server"))
async def server(message: Message) -> None:
    telegram_id = message.from_user.id

    if not await repository.user_exists(telegram_id):
        await message.answer(
            "Сначала зарегистрируйтесь командой /start."
        )
        return

    access = await repository.get_user_access(telegram_id)

    if access is None:
        access = await repository.get_free_access()

        if access is None:
            await message.answer(
                "Свободных серверов сейчас нет. Обратитесь к преподавателю"
            )
            return

        await repository.assign_access(
            access.id,
            telegram_id,
        )

        access = await repository.get_user_access(
            telegram_id
        )

    await message.answer(
        "Ваш сервер:\n\n"
        f"Домен: <code>{access.domain}</code>\n"
        f"Порт: <code>{access.port}</code>\n"
        f"Логин: <code>{access.login}</code>\n"
        f"Пароль: <code>{access.password}</code>\n\n"
        "Подключение:\n"
        f"<code>ssh {access.login}@{access.domain} -p {access.port}</code>"
    )
