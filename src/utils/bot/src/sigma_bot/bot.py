from aiogram import Bot, Dispatcher

from .config import config

bot = Bot(config.bot_token)
dp = Dispatcher()
