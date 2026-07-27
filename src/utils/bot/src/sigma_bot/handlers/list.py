from aiogram.filters import Command
from aiogram.types import Message, CallbackQuery

from ..keyboards.list import list_keyboard
from ..database.repository import repository

from .admin_router import admin_router


@admin_router.message(Command("list"))
async def list_command(message: Message) -> None:
    await message.answer(
        "---[ Выберите список сущностей ]---",
        reply_markup=list_keyboard(),
    )

@admin_router.callback_query(lambda callback: callback.data == "list:servers")
async def list_servers(callback: CallbackQuery) -> None:
    accesses = await repository.get_all_accesses()

    if not accesses:
        await callback.answer("Нет данных", show_alert=True)
        return

    lines = [
        "<b>Список серверов</b>",
        "",
        "<code>№  Login               User</code>",
    ]

    for index, access in enumerate(accesses, start=1):
        owner = f"@{access.username}" if access.username else "-"

        lines.append(
            f"<code>{index:02} "
            f"{access.login:<22} "
            f"{owner}</code>"
        )

    await callback.message.edit_text(
        "\n".join(lines),
        reply_markup=list_keyboard(),
    )

    await callback.answer()

@admin_router.callback_query(lambda callback: callback.data == "list:users")
async def list_users(callback: CallbackQuery) -> None:
    users = await repository.get_all_users()

    if not users:
        await callback.answer(
            "Нет зарегистрированных пользователей.",
            show_alert=True,
        )
        return

    lines = [
        "<b>Список пользователей</b>",
        "",
    ]

    for index, user in enumerate(users, start=1):
        username = (
            f"@{user.username}"
            if user.username
            else "-"
        )

        server = user.login or "-"

        lines.append(
            f"{index:02}. "
            f"{username}\n"
            f"    {user.full_name}\n"
            f"    Сервер: {server}\n"
        )

    await callback.message.edit_text(
        "\n".join(lines),
        reply_markup=list_keyboard(),
    )

    await callback.answer()
