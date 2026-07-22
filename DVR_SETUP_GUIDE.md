# CP Plus DVR (Dahua OEM) — Stream Access & Capabilities Guide

**Device:** CP-UVR-0801E1-CS (8-Channel DVR)  
**Firmware:** V3.218.00AT005.2  
**Web GUI:** v3.2.7.84189  
**IP Address:** `192.168.1.240`  
**Credentials:** `admin` / `zmjjkk77`  

---

## 1. How to Pull Video Streams (RTSP)

The DVR acts as an **RTSP server**. You pull streams from it using the following URL structure:

```
rtsp://admin:zmjjkk77@192.168.1.240:554/cam/realmonitor?channel=<CH>&subtype=0
```

* `<CH>` = Channel number (`1` through `8`)
* `subtype=0` = Main Stream (high resolution)

### Channel Breakdown

| Channel | Resolution | FPS | Aspect Ratio | Bitrate | Status |
|:-------:|:----------:|:---:|:------------:|:-------:|:------:|
| **1** | 960×1080 | 30 | 16:9 (SAR 2:1) | ~1.4 Mbps | Active |
| **2** | 960×1080 | 24 | 16:9 (SAR 2:1) | ~1.3 Mbps | Active |
| **3** | 960×1080 | 24 | 16:9 (SAR 2:1) | ~1.3 Mbps | Active |
| **4** | 1280×720 | 30 | 16:9 | ~1.4 Mbps | Active |
| **5** | 960×1080 | 24 | 16:9 (SAR 2:1) | ~1.3 Mbps | Active |
| **6** | 960×1080 | 24 | 16:9 (SAR 2:1) | ~1.3 Mbps | Active |
| **7** | 960×1080 | 24 | 16:9 (SAR 2:1) | ~1.3 Mbps | Active |
| **8** | 960×1080 | 25 | 16:9 (SAR 2:1) | ~1.4 Mbps | Active |

> **Note:** Sub-streams (`subtype=1`) are disabled/unavailable on these cameras.

---

## 2. Ingestion Examples

### A. FFmpeg (Command Line)

**Single Channel (CLI test):**
```bash
ffmpeg -rtsp_transport tcp \
  -i "rtsp://admin:zmjjkk77@192.168.1.240:554/cam/realmonitor?channel=1&subtype=0" \
  -c copy -f flv "rtmp://your-server/live/cam1"
```

**Record to MP4 file:**
```bash
ffmpeg -rtsp_transport tcp \
  -i "rtsp://admin:zmjjkk77@192.168.1.240:554/cam/realmonitor?channel=1&subtype=0" \
  -t 60 -c copy output_ch1.mp4
```

**Forward All 8 Channels (Background jobs):**
```bash
#!/bin/bash
for ch in {1..8}; do
  ffmpeg -rtsp_transport tcp \
    -i "rtsp://admin:zmjjkk77@192.168.1.240:554/cam/realmonitor?channel=${ch}&subtype=0" \
    -c copy -f flv "rtmp://your-server/live/cam${ch}" &
done
wait
```

### B. MediaMTX / go2rtc (Configuration File)

Add to `mediamtx.yml`:

```yaml
paths:
  cam1:
    source: rtsp://admin:zmjjkk77@192.168.1.240:554/cam/realmonitor?channel=1&subtype=0
  cam2:
    source: rtsp://admin:zmjjkk77@192.168.1.240:554/cam/realmonitor?channel=2&subtype=0
  cam3:
    source: rtsp://admin:zmjjkk77@192.168.1.240:554/cam/realmonitor?channel=3&subtype=0
  cam4:
    source: rtsp://admin:zmjjkk77@192.168.1.240:554/cam/realmonitor?channel=4&subtype=0
  cam5:
    source: rtsp://admin:zmjjkk77@192.168.1.240:554/cam/realmonitor?channel=5&subtype=0
  cam6:
    source: rtsp://admin:zmjjkk77@192.168.1.240:554/cam/realmonitor?channel=6&subtype=0
  cam7:
    source: rtsp://admin:zmjjkk77@192.168.1.240:554/cam/realmonitor?channel=7&subtype=0
  cam8:
    source: rtsp://admin:zmjjkk77@192.168.1.240:554/cam/realmonitor?channel=8&subtype=0
```

### C. VLC Media Player

1. Go to **Media > Open Network Stream** (`Ctrl+N`)
2. Enter: `rtsp://admin:zmjjkk77@192.168.1.240:554/cam/realmonitor?channel=1&subtype=0`
3. Click **Play**

### D. Python (OpenCV)

```python
import cv2

url = "rtsp://admin:zmjjkk77@192.168.1.240:554/cam/realmonitor?channel=1&subtype=0"
cap = cv2.VideoCapture(url)

while cap.isOpened():
    ret, frame = cap.read()
    if not ret:
        break
    cv2.imshow('DVR Channel 1', frame)
    if cv2.waitKey(1) & 0xFF == ord('q'):
        break

cap.release()
cv2.destroyAllWindows()
```

---

## 3. Network Ports & Services

| Port | Service | Status | Description |
|:----:|:-------:|:------:|:------------|
| **80/tcp** | HTTP | OPEN | Web management UI + RPC2 JSON API |
| **554/tcp** | RTSP | OPEN | Video stream server (Digest auth) |
| **25001/tcp** | Dahua SDK | OPEN | Dahua private SDK protocol (SmartPSS/ConfigTool) |
| **123/udp** | NTP | OPEN | Time synchronization service |
| **443/tcp** | HTTPS | DISABLED | Disabled in configuration |
| **21, 22, 23** | FTP/SSH/Telnet | CLOSED | No hidden command-line remote access |
| **3702/udp** | ONVIF | ABSENT | ONVIF WS-Discovery is disabled/not present |

---

## 4. What Features & APIs Are Available

### A. JSON-RPC API (`/RPC2`)
The device exposes a full **JSON-RPC 2.0 API** on port 80. All configuration, system info, and user management go through this interface.

* **Login Endpoint:** `POST /RPC2_Login`
* **API Endpoint:** `POST /RPC2`

**System Info Method:**
```bash
curl -X POST http://192.168.1.240/RPC2 \
  -H "Content-Type: application/json" \
  -d '{"method":"magicBox.getSystemInfo","params":{},"session":"<SESSION_ID>","id":1}'
```

**Supported Methods:**
- `magicBox.getSystemInfo` — Model, serial number, hardware version
- `magicBox.getSerialNo` — Serial number (`2106011801007129`)
- `magicBox.getSoftwareVersion` — Software build version
- `magicBox.getVendor` — Vendor (`CPPLUS`)
- `configManager.getConfig` — Dump configuration categories (`Network`, `Video`, `Record`, `All`)
- `mediaFileFind.findFile` — Search recorded video on hard drives
- `global.keepAlive` — Maintain RPC session active

### B. Unauthenticated Info Endpoint
The device leaks internal capability details **without authentication**:

```bash
curl http://192.168.1.240/current_config/WebCapConfig
```

Returns:
* Serial Number: `2106011801007129`
* Default IP: `192.168.1.240`
* Vendor: `CPPLUS`
* Supported Login Types: TCP, UDP, Multicast
* Features: MQTT support, Cloud Update, PIR Alarm support

### C. Alarm & Event Pushing (HTTP / MQTT)
While the DVR **cannot push video**, it **can push alarm events** to an external server:
* **HTTP Alarm Server:** Can send HTTP POST notifications to a designated server URL when motion or IVS events trigger.
* **MQTT:** `b_supportMqtt: true` — can publish alarm state changes to an MQTT broker.

### D. Software Compatibility
* **SmartPSS:** Fully supported via TCP port 25001 (Dahua/CP Plus desktop VMS)
* **DMSS (Mobile App):** Supported via P2P / LAN discovery
* **CP Plus ConfigTool:** Supported via SDK port 25001

---

## 5. Limitations & Unsupported Features

| Feature | Supported? | Details / Explanation |
|:--------|:----------:|:----------------------|
| **RTMP Push** | ❌ No | DVR cannot initiate connections to RTMP servers |
| **RTSP Push** | ❌ No | DVR acts strictly as a server (pull-only) |
| **ONVIF** | ❌ No | Port 3702 / ONVIF endpoints return 404 |
| **CGI Endpoints** | ❌ No | Legacy `/configManager.cgi`, `/snapshot.cgi` return 404 |
| **Sub Streams** | ❌ No | `subtype=1` disabled on current camera configuration |
| **HTTPS (443)** | ❌ No | Port 443 closed |
| **Telnet / SSH** | ❌ No | Ports 22/23 closed |
| **Web Sockets** | ❌ No | No `/rtspoverwebsocket` support |

---

## 6. Security & Vulnerability Summary

1. **Authentication Bypass (CVE-2017-7921):**  
   Setting `Cookie: session=1` allows unauthenticated reads of `/current_config/WebCapConfig` and `/current_config/preLanguage`.

2. **CGI Endpoint Hardening:**  
   The firmware (V3.218.00AT005.2) has completely disabled older unauthenticated CGI scripts. All operations require RPC2 challenge-response auth.

3. **Rate Limiting / Lockout:**  
   Failing 5 consecutive login attempts triggers a **300-second (5-minute) IP lockout** (`remainLockSecond: 300`).

---

## 7. Quick Cheat Sheet

```bash
# Stream URL pattern
rtsp://admin:zmjjkk77@192.168.1.240:554/cam/realmonitor?channel=[1-8]&subtype=0

# Verify channel 1
ffprobe -rtsp_transport tcp "rtsp://admin:zmjjkk77@192.168.1.240:554/cam/realmonitor?channel=1&subtype=0"

# Record 30s sample
ffmpeg -rtsp_transport tcp -i "rtsp://admin:zmjjkk77@192.168.1.240:554/cam/realmonitor?channel=1&subtype=0" -t 30 -c copy sample.mp4

# Unauthenticated config leak
curl http://192.168.1.240/current_config/WebCapConfig
```
