FROM python:3.12-slim

WORKDIR /app

COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

COPY inference ./inference
COPY models ./models

EXPOSE 8090

CMD ["uvicorn", "inference.app:app", "--host", "0.0.0.0", "--port", "8090"]