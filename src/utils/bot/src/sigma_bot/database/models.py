from dataclasses import dataclass


@dataclass(slots=True)
class User:
    telegram_id: int
    username: str | None
    full_name: str
    registered_at: str


@dataclass(slots=True)
class Access:
    id: int
    login: str
    password: str
    port: int
    domain: str
    telegram_id: int | None


@dataclass(slots=True)
class AccessInfo:
    login: str
    password: str
    port: int
    username: str | None


@dataclass(slots=True)
class UserInfo:
    username: str | None
    full_name: str
    login: str | None


@dataclass(slots=True)
class UserDetails:
    telegram_id: int
    username: str | None
    full_name: str
    registered_at: str
    login: str | None
    password: str | None
    port: int | None
    domain: str | None
