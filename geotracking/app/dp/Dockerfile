FROM python:3.13-alpine AS build
RUN apk add gcc musl-dev
RUN pip install -U pip
RUN pip install uv
WORKDIR /src
ADD uv.lock .
ADD pyproject.toml .
RUN uv venv /opt
RUN . /opt/bin/activate \
&&  uv sync --no-install-project --active

FROM python:3.13-alpine
COPY --from=build /opt /opt
ENV PYTHONPATH /opt/lib/python3.13/site-packages
ENV PATH /opt/bin:${PATH}
RUN adduser --system --home /app app
WORKDIR /app
COPY . .
ENTRYPOINT [ "python", "main.py" ]
