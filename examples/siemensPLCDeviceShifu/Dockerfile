FROM python:slim

COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

COPY siemens-plc.py .

CMD ["python", "siemens-plc.py"] 
