from aiogram.filters import CommandStart
from aiogram.fsm.context import FSMContext
from aiogram.types import Message

from ..config import config
from ..database.repository import repository
from ..states.registration import Registration

from .user_router import user_router


@user_router.message(CommandStart())
async def start(
    message: Message,
    state: FSMContext,
) -> None:
    telegram_id = message.from_user.id

    if await repository.user_exists(telegram_id):
        await message.answer(
            "Вы уже зарегистрированы.\n\n"
            "Введите команду /server для получения данных подключения."
        )
        return

    await message.answer(
        "Добро пожаловать!\n\n"
        f"Вот ссылка на чат курса:\n{config.course_chat}"
    )

    registration_message = await message.answer(
        "Введите ваше ФИО."
    )

    await state.update_data(
        registration_message_id=registration_message.message_id
    )

    await state.set_state(
        Registration.waiting_for_full_name
    )


@user_router.message(Registration.waiting_for_full_name)
async def process_full_name(
    message: Message,
    state: FSMContext,
) -> None:
    full_name = message.text.strip()

    if len(full_name) < 5:
        await message.answer(
            "Пожалуйста, введите корректное ФИО."
        )
        return

    data = await state.get_data()

    await repository.create_user(
        telegram_id=message.from_user.id,
        username=message.from_user.username,
        full_name=full_name,
    )

    await message.bot.delete_message(
        chat_id=message.chat.id,
        message_id=data["registration_message_id"],
    )

    await message.delete()

    await message.answer(
        "Регистрация успешно завершена.\n\n"
        "Теперь введите команду /server для получения данных подключения."
    )

    await state.clear()
