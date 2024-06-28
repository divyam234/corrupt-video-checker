# Corrupt Video Checker

## Usage
```sh
corrupt-video-check -path  input.mp4 -interval 60

```
- `FFMPEG` and `FFPROBE` required and should be PATH
- Checks if remote video file is corrupted by  extracting frames at regular intervals.
- Frame extraction will fail if video is corrupted at any timestamp.
- Works for local and remote file.
- Default interval is 60 second you can decrease it to make it more accurate for local files.

## License
This project is licensed under the [MIT License](LICENSE).
