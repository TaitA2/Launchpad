import mido

with mido.open_input('Launchpad S') as inport:
    print("Listening for MIDI...")
    for msg in inport:
        print(msg)
