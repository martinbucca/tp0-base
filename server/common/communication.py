from common.utils import Bet

BET_MESSAGE_ID = "BET"
SEPARATOR = "|"
CHUNK_BET_MESSAGE_ID = 12
CHUNK_FINISH_MESSAGE_ID = 13
AGENCY_SUCCESS_MESSAGE_ID = 14
AGENCY_ERROR_MESSAGE_ID = 15
FINISH_MESSAGE_ID = 16



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

        fields = payload.decode("utf-8").split("|")
        if fields[0] == BET_MESSAGE_ID:
            fields = fields[1:]
            bet = Bet(*fields)
            return bet
        return None

    def deserialize_chunk(payload: bytes) -> (str, list[Bet]):
        fields = payload.decode("utf-8").split("&")
        client_id = fields[0]
        chunk_id = fields[1]
        bets = fields[2:]
        bets_list = []
        for bet in bets:
            bet_fields = bet.split("|")
            if len(bet_fields) == 5:
                name = bet_fields[0]
                surname = bet_fields[1]
                document_id = bet_fields[2]
                birth_date = bet_fields[3]
                number = bet_fields[4]
                bet = Bet(agency=client_id, first_name=name, last_name=surname, document=document_id, birthdate=birth_date, number=number)
                bets_list.append(bet)
        return (chunk_id, bets_list)

    def receive_bets_chunk(self):
        message_id_byte = self.socket.recv(1)
        if not message_id_byte:
            raise ConnectionError("Failed to read message ID")
        message_id = int.from_bytes(message_id_byte, byteorder='big')
        if message_id == CHUNK_BET_MESSAGE_ID:
            length_bytes = self.socket.recv(2)
            if len(length_bytes) < 2:
                raise ConnectionError("Failed to read message length")
            msg_length = int.from_bytes(length_bytes, byteorder='big')
            payload = b""
            while len(payload) < msg_length:
                chunk = self.socket.recv(msg_length - len(payload))
                if not chunk:
                    raise ConnectionError("Connection closed before receiving full message")
                payload += chunk
            chunk_id, bets_list = self.deserialize_chunk(payload)
            return (chunk_id, bets_list)

        elif message_id == CHUNK_FINISH_MESSAGE_ID:
            self.send_finish_message(client_id)
        return None

    def send_ok_message(self, chunk_id):
        message_id = AGENCY_SUCCESS_MESSAGE_ID.to_bytes(1, byteorder='big')
        chunk_id = chunk_id.to_bytes(4, byteorder='big')
        self.socket.sendall(message_id + chunk_id)

    def send_finish_message(self, client_id):
        message_id = FINISH_MESSAGE_ID.to_bytes(1, byteorder='big')
        client_id = client_id.to_bytes(4, byteorder='big')
        self.socket.sendall(message_id + client_id)


    def close(self):
        self.socket.close()

    def getpeername(self):
        return self.socket.getpeername()


    
