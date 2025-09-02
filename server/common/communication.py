from common.utils import Bet

class AgencySocket:
    def __init__(self, socket):
        self.socket = socket

    def receive_bet(self):
        # Read the first 4 bytes for the length of the message
        length_bytes = self.socket.recv(4)
        if len(length_bytes) < 4:
            raise ConnectionError("Failed to read message length")
        msg_length = int.from_bytes(length_bytes, byteorder='big')

        # Read the payload based on the length
        payload = b""
        while len(payload) < msg_length:
            chunk = self.socket.recv(msg_length - len(payload))
            if not chunk:
                raise ConnectionError("Connection closed before receiving full message")
            payload += chunk

        # Decode and split fields
        fields = payload.decode("utf-8").split("|")
        # Remove the first field (message id)
        fields = fields[1:]
        # Create Bet object with remaining fields
        bet = Bet(*fields)
        return bet

    def send_message(self, msg):
        msg_bytes = msg.encode("utf-8")
        length_prefix = len(msg_bytes).to_bytes(4, byteorder='big')
        self.socket.sendall(length_prefix + msg_bytes)

    def send_ok_message(self):
        self.send_message("OK")

    def send_error_message(self):
        self.send_message("ERROR")

    def close(self):
        self.socket.close()

    def getpeername(self):
        return self.socket.getpeername()


    
