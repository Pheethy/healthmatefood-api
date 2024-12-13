# เริ่มต้นจากหยิบ golang image มาเป็น base image
FROM golang:1.22-alpine
# ทำการกำหนด /go/src/flavorparser เป็น path เริ่มต้น
WORKDIR /go/src/healthmatefood-api
# Copy source ทั้งหมดจาก directory ปัจจุบันสู่ working directory ภายใน container
COPY . .
# ทำการ get complieDaemon สำหรับ run
RUN go get -u github.com/githubnemo/CompileDaemon
# ทำติดตั้ง complieDaemon สำหรับ run
RUN go install github.com/githubnemo/CompileDaemon
# download dependencies ทั้งหมดที่ใช้
RUN go mod tidy
RUN mkdir -p /go/src/healthmatefood-api/tmp
# Build the Go app.
RUN go build -o /go/src/healthmatefood-api/tmp/app main.go
# Expose port 8080 ออกมาภายนอก container
EXPOSE 8080
# Set environment variable for service account key file
ENV GOOGLE_APPLICATION_CREDENTIALS=/go/src/healthmatefood-api/asset/gcp.json
# กำหนดคำสั่งหลักที่จะรันเมื่อ container ถูกเรียกใช้งาน ในที่นี้คือเรียกใช้ compiledaemon
ENTRYPOINT CompileDaemon -include=go.mod -log-prefix=false -color=true -build="go build -o ./tmp/app main.go" -command="./tmp/app"
