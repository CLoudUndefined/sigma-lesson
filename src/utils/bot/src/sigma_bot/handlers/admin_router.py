from aiogram import Router

from ..filters.admin import IsAdmin


admin_router = Router()
admin_router.message.filter(IsAdmin())
