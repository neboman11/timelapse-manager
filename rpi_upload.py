import cv2 as cv
from datetime import datetime
from picamera2 import Picamera2, Preview
import requests
import sched
import sys
import time

picam2 = Picamera2()
current_id = 0

preview_config = picam2.create_preview_configuration(main={"size": (1280, 720)})
picam2.configure(preview_config)


def take_picture(scheduler):
    global current_id
    scheduler.enter(5, 1, take_picture, (scheduler,))

    try:
        time.sleep(2)

        picam2.capture_file("current_frame.jpg")

        image = cv.imread("current_frame.jpg")
        image = cv.flip(image, 0)
        image = cv.putText(
            image,
            datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
            (20, 20),
            cv.FONT_HERSHEY_SIMPLEX,
            2,
            255,
        )

        response = requests.post(
            f"http://timelapse/inprogress/add?id={current_id}",
            cv.imencode(".jpg", image)[1].tobytes(),
            headers={"Content-Type": "image/jpeg"},
        )

        response.raise_for_status()
        current_id = response.content.decode()
    except Exception as e:
        print(f"Failure during image upload: {e}")


def main():
    global current_id
    if len(sys.argv) > 1:
        current_id = sys.argv[1]
    picam2.start()
    my_scheduler = sched.scheduler(time.time, time.sleep)
    my_scheduler.enter(5, 1, take_picture, (my_scheduler,))
    my_scheduler.run()


if __name__ == "__main__":
    main()
