from dataclasses import dataclass
import os

from dotenv import load_dotenv

load_dotenv()


@dataclass(frozen=True)
class Config:
    bot_token: str
    admin_ids: set[int]
    course_chat: str

config = Config(
    bot_token=os.getenv("BOT_TOKEN", ""),
    admin_ids={
        int(admin_id)
        for admin_id in os.getenv("ADMIN_IDS", "").split(",")
        if admin_id
    },
    course_chat=os.getenv("COURSE_CHAT_LINK", ""),
)

