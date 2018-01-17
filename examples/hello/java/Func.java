import com.google.gson.Gson;
import com.google.gson.stream.JsonReader;

import java.io.InputStreamReader;

public class Func {

    /**
     * @param args the command line arguments
     */
    public static void main(String[] args) {
        JsonReader br = new JsonReader(new InputStreamReader(System.in));
        Gson gson = new Gson();
        Payload payload = gson.fromJson(br, Payload.class);
        if (payload == null) {
            payload = new Payload("world");
        }

        System.out.printf("Hello %s!\n", payload.getName());

    }
}


