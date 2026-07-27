from aiogram.filters import Command
from aiogram.types import Message

from ..database.repository import repository

from .user_router import user_router


@user_router.message(Command("forget_me"))
async def forget_me(message: Message) -> None:
    telegram_id = message.from_user.id

    if not await repository.user_exists(telegram_id):
        await message.answer(
            "Вы еще не зарегистрированы."
        )
        return

    access = await repository.get_user_access(telegram_id)

    if access is not None:
        await repository.release_access(access.id)

    await repository.delete_user(telegram_id)

    await message.answer(
        "Ваши данные успешно удалены.\n\n"
        "Теперь вы можете снова зарегистрироваться командой /start."
    )
