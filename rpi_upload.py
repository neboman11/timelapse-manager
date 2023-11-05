import sched
from picamera2 import Picamera2, Preview
import requests
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

        metadata = picam2.capture_file("current_frame.jpg")
        print(metadata)

        image_content = None
        with open("current_frame.jpg", "rb") as f:
            image_content = f.read()

        response = requests.post(
            f"http://timelapse/inprogress/add?id={current_id}", image_content
        )

        response.raise_for_status()
        current_id = response.content.decode()
    except Exception as e:
        print(f"Failure during image upload: {e}")


def main():
    picam2.start()
    my_scheduler = sched.scheduler(time.time, time.sleep)
    my_scheduler.enter(5, 1, take_picture, (my_scheduler,))
    my_scheduler.run()


if __name__ == "__main__":
    main()
