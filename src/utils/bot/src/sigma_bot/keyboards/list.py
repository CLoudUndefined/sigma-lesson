from aiogram.types import InlineKeyboardButton
from aiogram.utils.keyboard import InlineKeyboardBuilder


def list_keyboard() -> InlineKeyboardMarkup:
    builder = InlineKeyboardBuilder()

    builder.button(
        text="Серверы",
        callback_data="list:servers",
    )

    builder.button(
        text="Пользователи",
        callback_data="list:users",
    )

    builder.adjust(2)

    return builder.as_markup()
