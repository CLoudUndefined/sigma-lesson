import asyncio
import logging

logging.basicConfig(level=logging.INFO)

from aiogram import Bot
from aiogram.enums import ParseMode
from aiogram.client.default import DefaultBotProperties

from .bot import dp
from .config import config
from .database.db import database

from .handlers.user_router import user_router
from .handlers.admin_router import admin_router


async def main():
    await database.connect()

    await database.initialize()

    bot = Bot(
        token=config.bot_token,
        default=DefaultBotProperties(parse_mode=ParseMode.HTML),
    )

    dp.include_router(user_router)
    dp.include_router(admin_router)

    print("All systems authentificated")

    try:
        print("Starting polling")
        await dp.start_polling(bot)
        print("Polling returned")
    except Exception:
        logging.exception("Polling crashed")
        raise
    finally:
        print("Closing database")
        await database.close()

if __name__ == "__main__":
    asyncio.run(main())
