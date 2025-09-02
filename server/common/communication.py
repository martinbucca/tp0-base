

class AgencySocket:
    def __init__(self, socket):
        self.socket = socket

    def receive_bet(self):
        # Read the first 4 bytes for the length of the message
        length_bytes = self.socket.recv(4)
        if len(length_bytes) < 4:
            raise ConnectionError("Failed to read message length")
        msg_length = int.from_bytes(length_bytes, byteorder='big')

        # Read the rest of the message based on the length
        msg = b""
        while len(msg) < msg_length:
            chunk = self.socket.recv(msg_length - len(msg))
            if not chunk:
                raise ConnectionError("Connection closed before receiving full message")
            msg += chunk
        return msg.decode("utf-8")

    
